package http

import (
	"database/sql"
	"github.com/gcjensen/splend-api"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

func GetUserMonthBreakdown(dbh *sql.DB) httprouter.Handle {
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

		month := params.ByName("month")

		user, err := splend.NewUserFromDB(id, dbh)
		var breakdown []splend.CategoryTotals
		if err == nil {
			breakdown, err = user.GetMonthBreakdown(month)
		}

		if err != nil {
			respondWithError(err, writer)
			return
		}

		respondWithJSON(writer, http.StatusOK, breakdown)
	})
}
