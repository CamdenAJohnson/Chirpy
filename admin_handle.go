package main

import "net/http"

// SetAdminRoutes Sets the routes for HTTP request starting with "/admin"
func SetAdminRoutes(router *http.ServeMux) {
	// Set the function handler for HTTP GET requests at /admin/metrics endpoint
	router.Handle("GET /admin/metrics", &APIConfig)

	// Set the function handler for HTTP POST reqeusts at /admin/reset endpoint
	router.Handle("POST /admin/reset", APIConfig.middlewareMetricsReset(http.HandlerFunc(resetHandler)))
}