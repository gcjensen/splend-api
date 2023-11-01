package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gcjensen/splend-api/splend"
)

func respondWithError(err error, writer http.ResponseWriter) {
	var code int

	var message string

	switch {
	case errors.Is(err, sql.ErrNoRows):
		code = http.StatusNotFound
		message = "User not found"
	case errors.Is(err, splend.ErrAlreadyExists):
		code = http.StatusBadRequest
		message = err.Error()
	default:
		code = http.StatusInternalServerError
		message = err.Error()
	}

	respondWithJSON(writer, code, map[string]string{"error": message})
}

func respondWithJSON(writer http.ResponseWriter, code int, resp interface{}) {
	response, _ := json.Marshal(resp)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)

	_, err := writer.Write(response)
	if err != nil {
		log.Printf("Failed to write response: %s", err.Error())
	}
}

func respondWithSuccess(writer http.ResponseWriter, code int, message string) {
	respondWithJSON(writer, code, map[string]string{"message": message})
}
