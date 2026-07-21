package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

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