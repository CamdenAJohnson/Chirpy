package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/CamdenAJohnson/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type serverConfig struct {
	Addr string
	Mux *http.ServeMux
	dbQueries *database.Queries
	Logger *log.Logger
}

// Global Variables
var ServerConfig serverConfig
var APIConfig apiConfig

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

// main it's the main function
func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil { log.Fatal(err) }
	
	// Set Server Settings
	ServerConfig = serverConfig{}
	ServerConfig.Addr = ":8080"
	ServerConfig.Mux = http.NewServeMux()
	ServerConfig.dbQueries = database.New(db)
	ServerConfig.Logger = log.Default()

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

	ServerConfig.Logger.Println("Starting on port :8080")
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		ServerConfig.Logger.Fatalf("Server error: %v\n", err)
	}
}