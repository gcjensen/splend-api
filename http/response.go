package http

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

func respondWithError(err error, writer http.ResponseWriter) {
	var code int
	var message string

	switch err {
	case sql.ErrNoRows:
		code = http.StatusNotFound
		message = "User not found"
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
	writer.Write(response)
}

func respondWithSuccess(writer http.ResponseWriter, code int, message string) {
	respondWithJSON(writer, code, map[string]string{"message": message})
}
