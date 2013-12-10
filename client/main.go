package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	url, err := url.Parse("http://localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	rp := httputil.NewSingleHostReverseProxy(url)

  log.Println("Starting server proxy on :2424")
  err = http.ListenAndServe(":2424", rp)
  if err != nil {
    log.Fatal(err)
  }
}
