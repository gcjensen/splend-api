package api

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gcjensen/amex"
	"github.com/gcjensen/splend-api/splend"
	"github.com/julienschmidt/httprouter"
)

const logDir = "/var/log/splend-api/"

func AddFromAmex(dbh *sql.DB) httprouter.Handle {
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

		decoder := json.NewDecoder(req.Body)

		var transaction amex.Transaction
		err = decoder.Decode(&transaction)

		if err != nil {
			respondWithError(err, writer)
			return
		}

		amexJSON, _ := json.Marshal(transaction)
		logTransaction("amex", amexJSON)

		err = user.AddAmexTransaction(transaction)
		if err != nil {
			log.Println(err.Error())
			respondWithError(err, writer)

			return
		}

		log.Printf("Transaction added from Amex")

		respondWithSuccess(writer, http.StatusOK, "Amex transaction added")
	}
}

func AddFromMonzo(dbh *sql.DB) httprouter.Handle {
	return func(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
		log.Printf("Transaction received from Monzo")

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

		decoder := json.NewDecoder(req.Body)

		var transaction map[string]interface{}
		err = decoder.Decode(&transaction)

		monzoJSON, _ := json.Marshal(transaction)
		logTransaction("monzo", monzoJSON)

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
			respondWithError(ErrUnregisteredWebhookType, writer)
			return
		}

		if err != nil {
			respondWithError(err, writer)
			return
		}

		respondWithSuccess(writer, http.StatusOK, "Request successful")
	}
}

func logTransaction(t string, txJSON []byte) {
	filename := t + "-" + time.Now().Format("2006-01-02 15:04:05") + ".json"
	_ = ioutil.WriteFile(logDir+filename, txJSON, 0o600)
}

// Checks the transaction is a debit i.e. negative and that the Monzo account is
// linked to the provided user.
func verifyTransaction(user *splend.User, data map[string]interface{}) bool {
	if merchant, ok := data["merchant"].(map[string]interface{}); ok {
		if _, ok := merchant["name"]; ok {
			return data["account_id"] == *user.MonzoAccount.ID &&
				data["amount"].(float64) < 0
		}
	}

	return false
}
