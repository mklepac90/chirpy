package main

import (
	"net/http";
	"sync/atomic";
	"fmt";
	"encoding/json";
	"log";
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) counterHandler (rw http.ResponseWriter,  req *http.Request) {
	rw.Header().Add("Content-Type", "text/html; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	counter := cfg.fileserverHits.Load()
	rw.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p><body></html>", counter)))
}

func (cfg *apiConfig) counterResetHandler (rw http.ResponseWriter,  req *http.Request) {
	rw.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
}

func healthzHandler (rw http.ResponseWriter,  req *http.Request) {
	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("OK"))
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func chirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, returnVals{
		Valid: true,
	})
}

func main() {
	mux := http.NewServeMux()
	apiCfg := apiConfig{}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("POST /api/validate_chirp", chirpHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.counterHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.counterResetHandler)

	srv := &http.Server{Addr: ":8080", Handler: mux}
	srv.ListenAndServe();
}