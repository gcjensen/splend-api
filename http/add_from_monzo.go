package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/gcjensen/splend-api"
	"github.com/julienschmidt/httprouter"
	"math"
	"net/http"
	"strconv"
)

func AddFromMonzo(dbh *sql.DB) httprouter.Handle {
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
			var transaction map[string]interface{}
			err = decoder.Decode(&transaction)

			if transaction["type"] == "transaction.created" {
				data := transaction["data"].(map[string]interface{})

				if verifyTransaction(user, data) {
					outgoing := &splend.Outgoing{
						Amount:      int(math.Abs(data["amount"].(float64))),
						Category:    "Other",
						Description: data["description"].(string),
						Spender:     *user.ID,
					}
					err = user.AddOutgoing(outgoing)
				}
			} else {
				err = errors.New("Unregistered webhook type")
				respondWithError(err, writer)
				return
			}
		}

		if err != nil {
			respondWithError(err, writer)
			return
		}

		respondWithSuccess(writer, http.StatusOK, "Request successful")
	})
}

/************************** Private Implementation ****************************/

/*
 * Checks the transaction is a debit i.e. negative and that the Monzo account
 * is linked to the provided user
 */
func verifyTransaction(user *splend.User, data map[string]interface{}) bool {
	return data["account_id"] == *user.LinkedAccounts.Monzo &&
		data["amount"].(float64) < 0
}
