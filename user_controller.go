package main

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type UserController struct {
	DB *sql.DB
}

func (c *UserController) GetDetails(
	writer http.ResponseWriter,
	req *http.Request,
) {
	// Pull out into some sort of reuable param verification logic
	params := mux.Vars(req)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		respondWithError(err, writer)
		return
	}

	user, err := NewFromDB(id, c.DB)
	if err != nil {
		respondWithError(err, writer)
		return
	}

	respondWithJSON(writer, http.StatusOK, user)
}

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
