package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
)

// Middleware functon defintion
type Middleware func(http.Handler) http.Handler

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