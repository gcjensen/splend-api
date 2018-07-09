package endpoints

import (
	"database/sql"
	"errors"
	"github.com/gcjensen/settle-api/outgoing"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

func DeleteOutgoing(dbh *sql.DB) httprouter.Handle {
	return httprouter.Handle(func(
		writer http.ResponseWriter,
		req *http.Request,
		params httprouter.Params,
	) {

		if id, err := strconv.Atoi(params.ByName("outgoingID")); err == nil {
			var o *outgoing.Outgoing
			if o, err = outgoing.New(id, dbh); err == nil {
				if err = o.Delete(); err == nil {
					msg := "Outgoing deleted!"
					respondWithSuccess(writer, http.StatusOK, msg)
					return
				} else {
					respondWithError(err, writer)
					return
				}
			}
		}

		respondWithError(errors.New("outgoingID parameter expected"), writer)
		return
	})
}
