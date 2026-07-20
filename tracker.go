package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// apiConfig sturct stores the number of times a request given the structs middleware functions is called
type apiConfig struct {
	fileserverHits atomic.Int32
}

// middlewareMetricsInc increases the count of APIConfig.fileserverHits
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		fmt.Printf("Hits: %v\n", cfg.fileserverHits.Load())
		next.ServeHTTP(w, r)
	})
}

// middlewareMetricsReset resets the count of APIConfig.fileserverHits
func (cfg *apiConfig) middlewareMetricsReset(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Store(0)
		next.ServeHTTP(w, r)
	})
}

// ServeHTTP serves HTML presenting APIConfig.fileserverHits
func (cfg *apiConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	num := cfg.fileserverHits.Load()
	w.Header().Set("Content-Type", "text/html charset=utf-8")

	str := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, num)

	fmt.Fprint(w, str)
}