package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	mux *http.ServeMux
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		fmt.Printf("Hits: %v\n", cfg.fileserverHits.Load())
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) middlewareMetricsReset(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Store(0)
		next.ServeHTTP(w, r)
	})
}

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

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Contorl-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}

func respondWithError(w http.ResponseWriter, code int, msg string) error {
	return respondWithJson(w, code, map[string]string{"error": msg})
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	rule := 140

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 400, "Something went wrong.")
		return
	}

	var chirp struct{
		Body string `json:"body"`
	}

	if err := json.Unmarshal(data, &chirp); err != nil {
		respondWithError(w, 400, "Something went wrong.")
		return
	}

	if len(chirp.Body) > rule {
		respondWithError(w, 400, "Body is to long.")
		return
	} else if len(chirp.Body) <= 0 {
		respondWithError(w, 400, "Empty Body.")
		return
	}

	cleaned := struct {
		Cleaned_body string `json:"cleaned_body"`
	}{
		Cleaned_body: "",
	}

	cleaned.Cleaned_body = cleanString(chirp.Body)

	respondWithJson(w, 200, cleaned)
}

func cleanString(str string) string {
	strArray := strings.Split(str, " ")
	var cleanArray []string

	for _, str := range strArray {
		strCopy := strings.ToLower(str)
		switch strCopy {
			case "kerfuffle":
				str = strings.ReplaceAll(strCopy, "kerfuffle", "****")
			case "sharbert":
				str = strings.ReplaceAll(strCopy, "sharbert", "****")
			case "fornax":
				str = strings.ReplaceAll(strCopy, "fornax", "****")
		}
		cleanArray = append(cleanArray, str)
	}
	return strings.Join(cleanArray, " ")
}

func main() {
	logger := log.Default()

	var apiCfg apiConfig
	apiCfg.mux = http.NewServeMux()

	buildMux(&apiCfg)

	httpServer := &http.Server{
		Addr: ":8080",
		Handler: apiCfg.mux,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Println("Starting on port :8080")
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Server error: %v\n", err)
	}
}

func buildMux(apiCfg *apiConfig) {
	mux := apiCfg.mux
	
	fsHandler := http.FileServer(http.Dir("."))

	metricsHandler := apiCfg.middlewareMetricsInc(fsHandler)

	mux.Handle("/app/", http.StripPrefix("/app", metricsHandler))

	mux.Handle("GET /admin/metrics", apiCfg)

	mux.Handle("POST /admin/reset", apiCfg.middlewareMetricsReset(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("POST /api/validate_chirp", validateChirp)
}