package http

import (
	"database/sql"
	"github.com/gcjensen/splend"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

func LogInUser(dbh *sql.DB) httprouter.Handle {
	return httprouter.Handle(func(
		writer http.ResponseWriter,
		req *http.Request,
		params httprouter.Params,
	) {

		// Pull out into some sort of reuable param verification logic
		id, err := strconv.Atoi(params.ByName("id"))
		if err != nil {
			respondWithError(err, writer)
			return
		}

		user, err := splend.NewUserFromDB(id, dbh)
		if err != nil {
			respondWithError(err, writer)
			return
		}

		respondWithJSON(writer, http.StatusOK, user)
	})
}