package http

import (
	"database/sql"
	"encoding/json"
	"github.com/gcjensen/splend-api"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

func AddOutgoing(dbh *sql.DB) httprouter.Handle {
	return httprouter.Handle(func(
		writer http.ResponseWriter,
		req *http.Request,
		params httprouter.Params,
	) {

		// Pull out into some sort of reuable param verification logic
		id, err := strconv.Atoi(params.ByName("id"))
		user, err := splend.NewUserFromDB(id, dbh)

		if err == nil {
			decoder := json.NewDecoder(req.Body)
			var outgoing splend.Outgoing
			err = decoder.Decode(&outgoing)

			if err == nil {
				err = user.AddOutgoing(&outgoing)
			}
		}

		if err != nil {
			respondWithError(err, writer)
			return
		}

		respondWithSuccess(writer, http.StatusCreated, "Outgoing added!")
	})
}
