package http

import (
	"database/sql"
	"encoding/json"
	"github.com/gcjensen/splend"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

func UpdateOutgoing(dbh *sql.DB) httprouter.Handle {
	return httprouter.Handle(func(
		writer http.ResponseWriter,
		req *http.Request,
		params httprouter.Params,
	) {

		id, err := strconv.Atoi(params.ByName("outgoingID"))
		outgoingToUpdate, err := splend.NewOutgoingFromDB(id, dbh)

		if err == nil {

			decoder := json.NewDecoder(req.Body)

			var updatedOutgoing splend.Outgoing
			err = decoder.Decode(&updatedOutgoing)

			if err == nil {

				outgoingToUpdate.Description = updatedOutgoing.Description
				outgoingToUpdate.Amount = updatedOutgoing.Amount
				outgoingToUpdate.Owed = updatedOutgoing.Owed
				outgoingToUpdate.Category = updatedOutgoing.Category

				err = outgoingToUpdate.Update()
			}
		}

		if err != nil {
			respondWithError(err, writer)
			return
		}

		respondWithSuccess(writer, http.StatusOK, "Outgoing updated!")
	})
}
