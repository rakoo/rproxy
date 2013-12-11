package main

import (
	"bytes"
	_ "compress/gzip"
	"crypto/sha1"
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

	rp := NewRProxy()
	proxy.OnRequest().Do(goproxy.FuncReqHandler(rp.addSignature))
	proxy.OnResponse().Do(goproxy.FuncRespHandler(rp.patch))

	http.Handle("/", proxy)
	log.Println("Starting listening on :2424")
	log.Fatal(http.ListenAndServe(":2424", nil))
}

type RProxy struct {
	cacheFile *os.File
}

func NewRProxy() *RProxy {
	f, err := os.Create("file")
	if err != nil {
		log.Println("Couldn't open file:", err)
	}
	return &RProxy{f}
}

func (rp *RProxy) HasCache() bool {
	return rp.cacheFile != nil && len(rp.ReadCache()) != 0
}
func (rp *RProxy) ReadCache() (content []byte) {
	_, err := rp.cacheFile.Seek(0, os.SEEK_SET)
	if err != nil {
		log.Println("Couldn't seek to beginning:", err)
		return
	}

	var b bytes.Buffer
	_, err = io.Copy(&b, rp.cacheFile)
	if err != nil {
		log.Println("Couldn't copy file to buffer:", err)
		return
	}
	return b.Bytes()
}
func (rp *RProxy) WriteCache(p []byte) {
	_, err := rp.cacheFile.Seek(0, os.SEEK_SET)
	if err != nil {
		log.Println("Couldn't seek to beginning:", err)
		return
	}

	rp.cacheFile.Write(p)
	rp.cacheFile.Sync()
}

func (rp *RProxy) addSignature(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	r.URL.Host = "localhost:8080"
	r.Header.Set("X-GoProxy", "yxorPoG-X")

	if !rp.HasCache() {
		return r, nil
	}

	sha := sha1.New()
	sha.Write(rp.ReadCache())
	log.Printf("Signing %x", sha.Sum(nil))

	var b bytes.Buffer
	enc := base64.NewEncoder(base64.URLEncoding, &b)

	rdiff := exec.Command("/usr/bin/rdiff", "signature", "-")
	rdiff.Stdin = bytes.NewReader(rp.ReadCache())
	rdiff.Stdout = enc

	err := rdiff.Run()
	if err != nil {
		log.Println("Error running rdiff signature:", err)
		return r, nil
	}

	// TODO: maybe wrap enc in a gzip.Writer. Has shown to actually take
	// more place, since signature should be small (<1024B)
	// If enc is a gzip.Writer, don't forget to flush it:
	// enc.Flush()

	log.Printf("Sending %d bytes as X-RProxy-Sig", b.Len())

	r.Header.Set("X-RProxy-Sig", string(b.Bytes()))
	return r, nil
}

func (rp *RProxy) patch(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	ct := r.Header.Get("Content-Type")
	required := "application/rproxy-patch"

	if !rp.HasCache() || ct != required && !strings.HasPrefix(ct, required+";") {
		var b bytes.Buffer
		io.Copy(&b, r.Body)
		rp.WriteCache(b.Bytes())

		r.Body = &bytesCloser{b}
		return r
	}

	defer r.Body.Close()

	var body bytes.Buffer
	_, err := io.Copy(&body, r.Body)
	if err != nil {
		log.Println("Couldn't copy body: ", err)
		return r
	}

	var patched bytes.Buffer

	rdiff := exec.Command("/usr/bin/rdiff", "patch", "file", "-", "-")
	rdiff.Stdin = &body
	rdiff.Stdout = &patched

	err = rdiff.Run()
	if err != nil {
		log.Println("Error running rdiff patch:", err)
	}

	rp.WriteCache(patched.Bytes())
	r.Body = &bytesCloser{patched}

	//log.Printf("Patch is %d bytes long; full content is %d bytes long", body.Len(), patched.Len())
	return r
}

type bytesCloser struct {
	bytes.Buffer
}

func (bc *bytesCloser) Close() error { return nil }
