package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// SetAPIRoutes Sets the routes for HTTP request starting with "/api"
func SetAPIRoutes(router *http.ServeMux) {
	// Set the function handler for HTTP GET requests at /api/healthz endpoint
	router.HandleFunc("GET /api/healthz", healthzHandle)

	// Set the function handler for HTTP POST reqeusts at /api/validate_chirp endpoint
	router.HandleFunc("POST /api/validate_chirp", validateChirp)

	// Set the function handler for HTTP POST requests at "/api/users endpoint"
	router.HandleFunc("POST /api/users", createUser)
}

// healthzHandle responeds with Status OK
func healthzHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

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