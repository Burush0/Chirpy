package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	fileHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", fileHandler)
	mux.HandleFunc("/healthz", healthCheck)
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
