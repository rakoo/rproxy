package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/elazarl/goproxy"
)

func main() {
	proxy := goproxy.NewProxyHttpServer()

	rp := NewRproxy()
	proxy.OnRequest().Do(goproxy.FuncReqHandler(rp.storeSig))
	proxy.OnResponse().Do(goproxy.FuncRespHandler(rp.delta))

	http.Handle("/", proxy)
	log.Println("Starting listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type RProxy struct {
	sig *os.File
}

func NewRproxy() *RProxy {
	sigFile, err := os.Create("sig")
	if err != nil {
		log.Println("Couldn't create sig:", err)
	}

	return &RProxy{sigFile}
}

func (rp *RProxy) HasSig() bool {
	return rp.sig != nil && len(rp.ReadSig()) != 0
}

func (rp *RProxy) ReadSig() (c []byte) {
	if rp.sig == nil {
		return
	}

	_, err := rp.sig.Seek(0, os.SEEK_SET)
	if err != nil {
		log.Println("Couldn't seek to beginning:", err)
		return
	}

	var b bytes.Buffer
	_, err = io.Copy(&b, rp.sig)
	if err != nil {
		log.Println("Couldn't copy to temp buff:", err)
		return
	}
	return b.Bytes()
}

func (rp *RProxy) WriteSig(p []byte) {
	if rp.sig == nil {
		return
	}

	_, err := rp.sig.Seek(0, os.SEEK_SET)
	if err != nil {
		log.Println("Couldn't seek to beginning:", err)
		return
	}

	rp.sig.Write(p)
	rp.sig.Sync()
}

func (rp *RProxy) storeSig(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	r.URL.Host = "localhost:9807"
	sig64 := r.Header.Get("X-RProxy-Sig")

	if sig64 == "" {
		return r, nil
	}

	log.Printf("C -> S: %dB", len(sig64))

	dec := base64.NewDecoder(base64.URLEncoding, strings.NewReader(sig64))
	var sig bytes.Buffer
	_, err := io.Copy(&sig, dec)
	if err != nil {
		log.Println("Couldn't decode signature:", err)
		return r, nil
	}
	rp.WriteSig(sig.Bytes())

	return r, nil
}

func (rp *RProxy) delta(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	if !rp.HasSig() {
		return r
	}

	r.Header.Set("Content-Type", "application/rproxy-patch")

	defer r.Body.Close()

	var body bytes.Buffer
	_, err := io.Copy(&body, r.Body)
	if err != nil {
		log.Println("Error recopying body: ", err)
		return r
	}

	var delta bytes.Buffer
	rdiff := exec.Command("/usr/bin/rdiff", "delta", "sig", "-", "-")
	rdiff.Stdin = &body
	rdiff.Stdout = &delta

	var errBuf bytes.Buffer
	rdiff.Stderr = &errBuf

	err = rdiff.Run()
	if err != nil {
		log.Println("Error running rdiff delta:", err)
		log.Println("rdiff error: ", string(errBuf.Bytes()))
	}

	r.Body = &bytesCloser{delta}

	log.Printf("S -> C: %dB", delta.Len())
	return r
}

type bytesCloser struct {
	bytes.Buffer
}

func (bc *bytesCloser) Close() error { return nil }
