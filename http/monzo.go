package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/gcjensen/splend-api"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

const logDir = "/var/log/splend-api/"

func AddFromMonzo(dbh *sql.DB) httprouter.Handle {
	return func(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
		id, err := strconv.Atoi(params.ByName("id"))
		user, err := splend.NewUserFromDB(id, dbh)

		log.Printf("Transaction received from Monzo")

		if err == nil {
			decoder := json.NewDecoder(req.Body)
			var transaction map[string]interface{}
			err = decoder.Decode(&transaction)

			monzoJson, _ := json.Marshal(transaction)
			filename := time.Now().Format("2006-01-02 15:04:05") + ".json"
			_ = ioutil.WriteFile(logDir+filename, monzoJson, 0644)

			if transaction["type"] == "transaction.created" {
				data := transaction["data"].(map[string]interface{})

				if verifyTransaction(user, data) {
					merchant := data["merchant"].(map[string]interface{})
					outgoing := &splend.Outgoing{
						Amount:      int(math.Abs(data["amount"].(float64))),
						Category:    "Other",
						Description: merchant["name"].(string),
						Spender:     *user.ID,
					}
					err = user.AddOutgoing(outgoing)
				} else {
					log.Printf("Transaction not valid. Ignoring")
				}
			} else {
				err = errors.New("unregistered webhook type")
				respondWithError(err, writer)
				return
			}
		}

		if err != nil {
			respondWithError(err, writer)
			return
		}

		respondWithSuccess(writer, http.StatusOK, "Request successful")
	}
}

/************************** Private Implementation ****************************/

/*
 * Checks the transaction is a debit i.e. negative and that the Monzo account
 * is linked to the provided user
 */
func verifyTransaction(user *splend.User, data map[string]interface{}) bool {
	if merchant, ok := data["merchant"].(map[string]interface{}); ok {
		if _, ok := merchant["name"]; ok {
			return data["account_id"] == *user.LinkedAccounts.Monzo &&
				data["amount"].(float64) < 0
		}
	}

	return false
}
