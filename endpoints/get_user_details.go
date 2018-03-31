package endpoints

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gcjensen/settle-api/user"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

func GetUserDetails(dbh *sql.DB) httprouter.Handle {
	return httprouter.Handle(func(
		writer http.ResponseWriter,
		req *http.Request,
		params httprouter.Params,
	) {

		// Pull out into some sort of reuable param verification logic
		id, err := strconv.Atoi(params.ByName("id"))
		if err != nil {
			fmt.Println(err)
			respondWithError(err, writer)
			return
		}

		user, err := user.NewFromDB(id, dbh)
		if err != nil {
			fmt.Println(err)
			respondWithError(err, writer)
			return
		}

		respondWithJSON(writer, http.StatusOK, user)
	})
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
