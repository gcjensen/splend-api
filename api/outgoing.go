package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gcjensen/splend-api/splend"
	"github.com/julienschmidt/httprouter"
)

func AddOutgoing(dbh *sql.DB) httprouter.Handle {
	return func(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
		id, err := strconv.Atoi(params.ByName("id"))
		if err != nil {
			respondWithError(err, writer)
			return
		}

		user, err := splend.NewUserFromDB(id, dbh)

		if err == nil {
			decoder := json.NewDecoder(req.Body)

			var outgoing splend.Outgoing
			err = decoder.Decode(&outgoing)

			if err == nil {
				err = user.AddOutgoing(&outgoing)
			}
		}

		if err != nil {
			respondWithError(err, writer)
			return
		}

		log.Printf("Outgoing added!")

		respondWithSuccess(writer, http.StatusCreated, "Outgoing added!")
	}
}

func DeleteOutgoing(dbh *sql.DB) httprouter.Handle {
	return func(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
		if id, err := strconv.Atoi(params.ByName("outgoingID")); err == nil {
			if o, err := splend.NewOutgoingFromDB(id, dbh); err == nil {
				if err = o.Delete(); err == nil {
					msg := "Outgoing deleted!"
					respondWithSuccess(writer, http.StatusOK, msg)

					return
				}

				respondWithError(err, writer)

				return
			}
		}

		respondWithError(ErrOutgoingIDExpected, writer)
	}
}

func SettleOutgoing(dbh *sql.DB) httprouter.Handle {
	return func(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
		id, err := strconv.Atoi(params.ByName("outgoingID"))
		if err != nil {
			respondWithError(err, writer)
			return
		}

		outgoing, err := splend.NewOutgoingFromDB(id, dbh)

		var shouldSettle int
		if err == nil {
			shouldSettle, err = strconv.Atoi(params.ByName("shouldSettle"))
			if err != nil {
				respondWithError(err, writer)
				return
			}

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
	}
}

func UpdateOutgoing(dbh *sql.DB) httprouter.Handle {
	return func(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
		id, err := strconv.Atoi(params.ByName("outgoingID"))
		if err != nil {
			respondWithError(err, writer)
			return
		}

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

				if err := outgoingToUpdate.Update(); err != nil {
					respondWithError(err, writer)
					return
				}

				err = outgoingToUpdate.UpdateTags(updatedOutgoing.Tags)
			}
		}

		if err != nil {
			respondWithError(err, writer)
			return
		}

		respondWithSuccess(writer, http.StatusOK, "Outgoing updated!")
	}
}
