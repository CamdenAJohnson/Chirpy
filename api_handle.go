package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/CamdenAJohnson/Chirpy/internal/auth"
	"github.com/CamdenAJohnson/Chirpy/internal/database"
	"github.com/google/uuid"
)

// SetAPIRoutes Sets the routes for HTTP request starting with "/api"
func SetAPIRoutes(router *http.ServeMux) {
	// Set the function handler for HTTP GET requests at /api/healthz endpoint
	router.HandleFunc("GET /api/healthz", healthzHandle)

	// Set the function handler for HTTP GET reqeusts at /api/chirps/ endpoint
	router.HandleFunc("GET /api/chirps", getChirps)

	// Set the function handler for HTTP GET requests at /api/chirps/{chirpid} endpoint
	router.HandleFunc("GET /api/chirps/{chirpid}", getChirpsById)

	// Set the function handler for HTTP POST requests at /api/chirps endpoint
	router.HandleFunc("POST /api/chirps", createChirp)

	// Set the function handler for HTTP POST requests at "/api/users endpoint"
	router.HandleFunc("POST /api/users", createUser)

	// Set the function handler for HTTP POST request at "/api/login endpoint"
	router.HandleFunc("POST /api/login", loginUser)
}

// healthzHandle responeds with Status OK
func healthzHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// createUser inserts a new user into the database
func createUser(w http.ResponseWriter, r *http.Request) {
	var requestFields struct{
		Email string `json:"email"`
		Password string `json:"password"`		
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to read request body.")
		return
	}

	if err := json.Unmarshal(data, &requestFields); err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to parse request body.")
		return
	}

	if len(requestFields.Email) <= 0  || len(requestFields.Password) <= 0{
		respondWithError(w, http.StatusBadRequest, "Bad Request: Email or Password not set.")
		return
	}

	if !strings.Contains(requestFields.Email, "@") && !strings.Contains(requestFields.Email, ".") {
		respondWithError(w, http.StatusBadRequest, "Bad Request: invalid email.")
		return
	}

	hashed_password, err := auth.HashPassword(requestFields.Password)
	if err != nil || len(hashed_password) <= 0 {
		ServerConfig.Logger.Printf("Failed to has password: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password.")
		return
	}

	dbParams := database.CreateUserParams{
		Email: requestFields.Email,
		HashedPassword: hashed_password,
	}

	newUser, err := ServerConfig.dbQueries.CreateUser(r.Context(), dbParams)
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

// loginUser 
func loginUser(w http.ResponseWriter, r *http.Request) {
	var requestFields struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to read request body.")
		return
	}

	if err := json.Unmarshal(data, &requestFields); err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to parse request body.")
		return
	}

	user, err := ServerConfig.dbQueries.GetUserByEmail(r.Context(), requestFields.Email)
	if err != nil {
		respondWithError(w, http.StatusNoContent, "No user linked with email.")
		return
	}

	hashed_password, err := auth.CheckPasswordHash(requestFields.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "")
		return
	}

	if !hashed_password {
		respondWithJson(w, http.StatusUnauthorized, "Incorrect password.")
		return
	}

	responsePayload := make(map[string]interface{})
	responsePayload["id"] = user.ID
	responsePayload["created_at"] = user.CreatedAt
	responsePayload["updated_at"] = user.UpdatedAt
	responsePayload["email"] = user.Email

	if err := respondWithJson(w, http.StatusOK, responsePayload); err != nil {
		ServerConfig.Logger.Printf("response helper failed to execute: %v", err)
	}
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

func getChirps(w http.ResponseWriter, r *http.Request) {
	var chirps struct {
		Chirp []struct {
			Id uuid.UUID `json:"id"`
			Created_at time.Time `json:"created_at"`
			Updated_at time.Time `json:"updated_at"`
			Body string `json:"body"`
			User_id uuid.UUID `json:"user_id"`
		}
	}

	chirpsData, err := ServerConfig.dbQueries.GetChirps(r.Context())
	if err != nil {
		ServerConfig.Logger.Printf("Error: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to execute querie.")
		return
	}

	for _, v := range chirpsData {
		chirp := struct {
			Id uuid.UUID `json:"id"`
			Created_at time.Time `json:"created_at"`
			Updated_at time.Time `json:"updated_at"`
			Body string `json:"body"`
			User_id uuid.UUID `json:"user_id"`
		}{
			Id: v.ID,
			Created_at: v.CreatedAt,
			Updated_at: v.UpdatedAt,
			Body: v.Body,
			User_id: v.UserID,
		}

		chirps.Chirp = append(chirps.Chirp, chirp)
	}

	if err := respondWithJson(w, http.StatusOK, chirps.Chirp); err != nil {
		ServerConfig.Logger.Printf("Error: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Parser Failed.")
	}
}

func getChirpsById(w http.ResponseWriter, r *http.Request) {
	var responseData struct {
		Id uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body string `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	chirpId := r.PathValue("chirpid")
	parsedUUID, err := uuid.Parse(chirpId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to parse chirp id.")
		return
	}

	chirp, err := ServerConfig.dbQueries.GetChirpsById(r.Context(), parsedUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound,  "Chirp does not exist.")
		return
	}

	responseData.Id = chirp.ID
	responseData.CreatedAt = chirp.CreatedAt
	responseData.UpdatedAt = chirp.UpdatedAt
	responseData.Body = chirp.Body
	responseData.UserId = chirp.UserID

	respondWithJson(w, http.StatusOK, responseData)
}