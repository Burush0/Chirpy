package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	fileHandler := http.FileServer(http.Dir("."))
	mux.Handle("/", fileHandler)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	log.Println("server is listening on port 8080")
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
