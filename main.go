package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/setToken", setTokens)
	http.HandleFunc("/refreshTokens", refreshTokensFunk)

	log.Println("server started on port 80")
	http.ListenAndServe(":80", nil)
}
