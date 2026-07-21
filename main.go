package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
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