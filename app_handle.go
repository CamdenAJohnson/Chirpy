package main

import "net/http"

// SetAppRoutes Sets the routes for HTTP requset starting with "/app"
func SetAppRoutes(router *http.ServeMux) {
	// Set up http.FileServer to let it know what dir to look at	
	fsHandler := http.FileServer(http.Dir("."))

	// Add apiConfig middlewareMetricsInc to track the number of request served
	metricsHandler := APIConfig.middlewareMetricsInc(fsHandler)

	// Set the function handler for HTTP GET requests at /app/ -> / endpoint
	router.Handle("/app/", http.StripPrefix("/app", metricsHandler))
}