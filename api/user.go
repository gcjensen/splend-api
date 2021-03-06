package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gcjensen/splend-api/splend"
	"github.com/julienschmidt/httprouter"
)

// How many months worth of outgoings to fetch.
const outgoingsMonths = 3

func GetUserMonthBreakdown(dbh *sql.DB) httprouter.Handle {
	return func(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
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
	}
}

func GetUserOutgoings(dbh *sql.DB) httprouter.Handle {
	return func(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
		id, err := strconv.Atoi(params.ByName("id"))
		if err != nil {
			respondWithError(err, writer)
			return
		}

		user, err := splend.NewUserFromDB(id, dbh)

		var outgoings []splend.Outgoing

		if err == nil {
			where := map[string]interface{}{"months": outgoingsMonths}
			outgoings, err = user.GetOutgoings(where)
		}

		if err != nil {
			respondWithError(err, writer)
			return
		}

		respondWithJSON(writer, http.StatusOK, outgoings)
	}
}

func GetUserSummary(dbh *sql.DB) httprouter.Handle {
	return func(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
		id, err := strconv.Atoi(params.ByName("id"))
		if err != nil {
			respondWithError(err, writer)
			return
		}

		user, err := splend.NewUserFromDB(id, dbh)

		var summary *splend.Summary
		if err == nil {
			summary, err = user.GetSummary()
		}

		if err != nil {
			respondWithError(err, writer)
			return
		}

		respondWithJSON(writer, http.StatusOK, summary)
	}
}

func LogInUser(dbh *sql.DB) httprouter.Handle {
	return func(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
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
	}
}
