package main

import (
	"net/http"
)

func HealthzHandler (rw http.ResponseWriter,  req *http.Request) {
	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("OK"))
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.Handle("/app/assets/logo.png", http.StripPrefix("/app/", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", HealthzHandler)

	srv := &http.Server{Addr: ":8080", Handler: mux}
	srv.ListenAndServe();
}