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
		couple, _ := strconv.Atoi(req.URL.Query().Get("couple"))

		user, err := splend.NewUserFromDB(id, dbh)
		var breakdown []splend.CategoryTotal
		if err == nil {
			breakdown, err = user.GetMonthBreakdown(month, couple == 1)
		}

		if err != nil {
			respondWithError(err, writer)
			return
		}

		respondWithJSON(writer, http.StatusOK, breakdown)
	})
}
