package api

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gcjensen/amex"
	"github.com/gcjensen/splend-api/splend"
	"github.com/julienschmidt/httprouter"
)

const logDir = "/tmp/log/splend-api/"

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
		if err := logTransaction("amex", amexJSON); err != nil {
			log.Printf("Error logging AMEX transaction: %s", err.Error())
		}

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
			log.Printf("Error parsing user ID: %s", err.Error())
			respondWithError(err, writer)
			return
		}

		user, err := splend.NewUserFromDB(id, dbh)
		if err != nil {
			log.Printf("Error finding user from ID: %s", err.Error())
			respondWithError(err, writer)
			return
		}

		decoder := json.NewDecoder(req.Body)

		var transaction map[string]interface{}
		err = decoder.Decode(&transaction)
		if err != nil {
			log.Printf("Error decoding transaction: %s", err.Error())
			respondWithError(err, writer)
			return
		}

		monzoJSON, _ := json.Marshal(transaction)
		if err := logTransaction("monzo", monzoJSON); err != nil {
			log.Printf("Error logging Monzo transaction: %s", err.Error())
		}

		if transaction["type"] == "transaction.created" {
			data := transaction["data"].(map[string]interface{})

			if valid, description := verifyTransaction(user, data); valid {
				outgoing := &splend.Outgoing{
					Amount:      int(math.Abs(data["amount"].(float64))),
					Category:    "Other",
					Description: description,
					Spender:     *user.ID,
				}

				if err = user.AddOutgoing(outgoing); err != nil {
					log.Printf("Error adding outgoing: %s", err.Error())
					respondWithError(err, writer)
					return
				}

				log.Printf("Outgoing added from Monzo")
			} else {
				log.Printf("Transaction not valid. Ignoring")
			}
		} else {
			log.Printf("%s: %s", ErrUnregisteredWebhookType.Error(), transaction["type"])
		}

		respondWithSuccess(writer, http.StatusOK, "Request successful")
	}
}

func logTransaction(t string, txJSON []byte) error {
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return err
	}

	filename := t + "-" + time.Now().Format("2006-01-02 15:04:05") + ".json"
	return ioutil.WriteFile(logDir+filename, txJSON, 0o600)
}

// Checks the transaction is a debit i.e. negative and that the Monzo account is
// linked to the provided user.
func verifyTransaction(user *splend.User, data map[string]interface{}) (bool, string) {
	var description string

	if merchant, ok := data["merchant"].(map[string]interface{}); ok {
		description, _ = merchant["name"].(string)
	} else if counterparty, ok := data["counterparty"].(map[string]interface{}); ok {
		description, _ = counterparty["name"].(string)
	} else {
		// No merchant or counterparty, so we're not interested
		return false, ""
	}

	accLinked := false
	for _, acc := range user.MonzoAccounts {
		if data["account_id"] == *acc.ID {
			accLinked = true
			break
		}
	}

	isDebit := data["amount"].(float64) < 0

	return accLinked && isDebit, description
}
