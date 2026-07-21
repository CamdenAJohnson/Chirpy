package main

import (
	"net/http"
	"os"
)

// SetAdminRoutes Sets the routes for HTTP request starting with "/admin"
func SetAdminRoutes(router *http.ServeMux) {
	// Set the function handler for HTTP GET requests at /admin/metrics endpoint
	router.Handle("GET /admin/metrics", &APIConfig)

	rsHandle := http.HandlerFunc(resetHandler)
	rsChain := CreateMiddlewareChain(middlewareCheckPlatform, APIConfig.middlewareMetricsReset)
	// Set the function handler for HTTP POST reqeusts at /admin/reset endpoint
	router.Handle("POST /admin/reset", rsChain(rsHandle))
}

// middlewareCheckPlatform checks the current env to either allow or reject request.
func middlewareCheckPlatform(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		platform := os.Getenv("PLATFORM")
		perm := os.Getenv("ADMIN")
		if platform != perm {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// resetHandler clears all stored data.
func resetHandler(w http.ResponseWriter, r *http.Request) {
	if err := ServerConfig.dbQueries.ClearUsers(r.Context()); err != nil {
		ServerConfig.Logger.Printf("Failed to execture db querie: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}