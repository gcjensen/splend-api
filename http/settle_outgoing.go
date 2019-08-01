package http

import (
	"database/sql"
	"github.com/gcjensen/splend-api"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

func SettleOutgoing(dbh *sql.DB) httprouter.Handle {
	return httprouter.Handle(func(
		writer http.ResponseWriter,
		req *http.Request,
		params httprouter.Params,
	) {

		// Pull out into some sort of reuable param verification logic
		id, err := strconv.Atoi(params.ByName("outgoingID"))
		outgoing, err := splend.NewOutgoingFromDB(id, dbh)

		var shouldSettle int
		if err == nil {
			shouldSettle, err = strconv.Atoi(params.ByName("shouldSettle"))
			err = outgoing.ToggleSettled(shouldSettle != 0)
		}

		if err != nil {
			respondWithError(err, writer)
			return
		}

		var message string
		if shouldSettle != 0 {
			message = "Outgoing settled!"
		} else {
			message = "Outgoing un-settled!"
		}

		respondWithSuccess(writer, http.StatusOK, message)
	})
}
