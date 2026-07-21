package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/CamdenAJohnson/Chirpy/internal/database"
	"github.com/google/uuid"
)

// SetAPIRoutes Sets the routes for HTTP request starting with "/api"
func SetAPIRoutes(router *http.ServeMux) {
	// Set the function handler for HTTP GET requests at /api/healthz endpoint
	router.HandleFunc("GET /api/healthz", healthzHandle)

	// Set the function handler for HTTP POST reqeusts at /api/validate_chirp endpoint
	//router.HandleFunc("POST /api/validate_chirp", validateChirp)

	// Set the function handler for HTTP POST requests at /api/chirps endpoint
	router.HandleFunc("POST /api/chirps", createChirp)

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

// createChirp validates the reqeust and inserts the data into the db
func createChirp(w http.ResponseWriter, r *http.Request) {
	maxLength := 140

	var requestFields struct{
		Body string `json:"body"`
		UserID string `json:"user_id"`
	}

	var chirpParams database.CreateChirpParams

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to read request body.")
		return
	}

	if err := json.Unmarshal(data, &requestFields); err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to parse request body.")
		return
	}

	if len(requestFields.Body) > maxLength {
		respondWithError(w, http.StatusBadRequest, "Bad Request: chirp is to long.")
		return
	}
	
	if len(requestFields.Body) <= 0 {
		respondWithError(w, http.StatusBadRequest, "Bad Request: Chirp can not be empty.")
		return
	}

	userId, err := uuid.Parse(requestFields.UserID)
	if err != nil {
		ServerConfig.Logger.Printf("Failed to parse user uuid: %v", err)
		respondWithError(w, http.StatusBadRequest, "Bad Request: Failed to parse userid.")
		return
	}

	chirpParams.Body = cleanString(requestFields.Body)
	chirpParams.UserID = userId
	chirp, err := ServerConfig.dbQueries.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		ServerConfig.Logger.Printf("Failed to execute db querie: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to execute querie.")
		return
	}

	responsePayload := make(map[string]interface{})
	responsePayload["id"] = chirp.ID
	responsePayload["created_at"] = chirp.CreatedAt
	responsePayload["updated_at"] = chirp.UpdatedAt
	responsePayload["body"] = chirp.Body
	responsePayload["user_id"] = chirp.UserID

	respondWithJson(w, http.StatusCreated, responsePayload)
}