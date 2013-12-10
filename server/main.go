package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	url, err := url.Parse("http://localhost:9807")
	if err != nil {
		log.Fatal(err)
	}

	rp := httputil.NewSingleHostReverseProxy(url)

  log.Println("Starting server proxy on :8080")
  err = http.ListenAndServe(":8080", rp)
  if err != nil {
    log.Fatal(err)
  }
}
