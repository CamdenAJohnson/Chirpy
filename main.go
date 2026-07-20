package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"
)

type serverConfig struct {
	Mux *http.ServeMux
	Addr string
}

// Global Variables
var ServerConfig serverConfig
var APIConfig apiConfig

// validateChirp checks the given http.Request body against a set of rules.
//
// Writing to http.ResponseWriter with the results
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

// cleanString replaces censored words. 
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

// Helper functions

// Middleware functon defintion
type Middleware func(http.Handler) http.Handler

// respondWithJson writes the given status code and json payload to http.ResponseWriter
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

// respondWithError writes the status code and message into http.ResponseWriter
func respondWithError(w http.ResponseWriter, code int, msg string) error {
	return respondWithJson(w, code, map[string]string{"error": msg})
}

// CreateMiddlewareChain creates a chain of middleware functions
func CreateMiddlewareChain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		// Iterate backwards to preserve the natural execution order
		for _, middleware := range slices.Backward(middlewares) {
			next = middleware(next)
		}
		return next
	}
} 

// Route Functions

// SetAppRoutes Sets the routes for HTTP requset starting with "/app"
func SetAppRoutes(router *http.ServeMux) {
	// Set up http.FileServer to let it know what dir to look at	
	fsHandler := http.FileServer(http.Dir("."))

	// Add apiConfig middlewareMetricsInc to track the number of request served
	metricsHandler := APIConfig.middlewareMetricsInc(fsHandler)

	// Set the function handler for HTTP GET requests at /app/ -> / endpoint
	router.Handle("/app/", http.StripPrefix("/app", metricsHandler))
}

// SetAdminRoutes Sets the routes for HTTP request starting with "/admin"
func SetAdminRoutes(router *http.ServeMux) {
	// Set the function handler for HTTP GET requests at /admin/metrics endpoint
	router.Handle("GET /admin/metrics", &APIConfig)

	// Set the function handler for HTTP POST reqeusts at /admin/reset endpoint
	router.Handle("POST /admin/reset", APIConfig.middlewareMetricsReset(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
}

// SetAPIRoutes Sets the routes for HTTP request starting with "/api"
func SetAPIRoutes(router *http.ServeMux) {
	// Set the function handler for HTTP GET requests at /api/healthz endpoint
	router.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Set the function handler for HTTP POST reqeusts at /api/validate_chirp endpoint
	router.HandleFunc("POST /api/validate_chirp", validateChirp)
}

// main it's the main function
func main() {
	logger := log.Default()
	
	// Set Server Settings
	ServerConfig.Addr = ":8080"
	ServerConfig.Mux = http.NewServeMux()

	// Configure the http.ServeMux
	SetAppRoutes(ServerConfig.Mux)
	SetAPIRoutes(ServerConfig.Mux)
	SetAdminRoutes(ServerConfig.Mux)

	httpServer := &http.Server{
		Addr: ServerConfig.Addr,
		Handler: ServerConfig.Mux,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Println("Starting on port :8080")
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Server error: %v\n", err)
	}
}