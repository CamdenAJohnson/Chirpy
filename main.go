package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
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

// createUser inserts a new user into the database
func createUser(w http.ResponseWriter, r *http.Request) {
	var requestFields struct{ Email string `json:"email"` }

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to read request body.")
		return
	}

	if err := json.Unmarshal(data, &requestFields); err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to parse request body.")
		return
	}

	if len(requestFields.Email) <= 0 {
		respondWithError(w, http.StatusBadRequest, "Bad Request: Email not set.")
		return
	}

	if !strings.Contains(requestFields.Email, "@") && !strings.Contains(requestFields.Email, ".") {
		respondWithError(w, http.StatusBadRequest, "Bad Request: invalid email.")
		return
	}

	newUser, err := ServerConfig.dbQueries.CreateUser(r.Context(), requestFields.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to execute db querie.")
		ServerConfig.Logger.Printf("Failed to execute 'CreateUser' querie: %v", err)
		return
	}

	responsePayload := make(map[string]interface{})
	responsePayload["id"] = newUser.ID
	responsePayload["created_at"] = newUser.CreatedAt
	responsePayload["updated_at"] = newUser.UpdatedAt
	responsePayload["email"] = newUser.Email

	if err := respondWithJson(w, http.StatusCreated, responsePayload); err != nil {
		ServerConfig.Logger.Printf("response helper failed to execute: %v", err)
	}
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
	router.Handle("POST /admin/reset", APIConfig.middlewareMetricsReset(http.HandlerFunc(resetHandler)))
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

	// Set the function handler for HTTP POST requests at "/api/users endpoint"
	router.HandleFunc("POST /api/users", createUser)
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