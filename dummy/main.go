package main

import (
  "log"
  "net/http"
)

var alt = false

func AlternateIndex(w http.ResponseWriter, r *http.Request) {
  if alt {
    http.ServeFile(w,r,"./index.html")
  } else {
    http.ServeFile(w,r,"./index2.html")
  }
  alt = !alt
}

func main() {
  http.HandleFunc("/", AlternateIndex)

  log.Println("Starting server")
  err := http.ListenAndServe(":9807", nil)
  if err != nil {
    log.Fatal(err)
  }
}
