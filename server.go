package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gcjensen/settle-api/endpoints"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	Router *httprouter.Router
	dbh    *sql.DB
}

func Auth(handler httprouter.Handle, dbh *sql.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		token := r.Header.Get("Token")
		// Sha256 the token and check against what we have stored for the user
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))

		id, err := strconv.Atoi(ps.ByName("id"))
		if err != nil {
			// Don't have the ID, so use the provided email to get it
			decoder := json.NewDecoder(r.Body)
			var body struct {
				Email string `json:"email"`
			}
			err = decoder.Decode(&body)

			if err == nil {
				statement := fmt.Sprintf(
					`SELECT id FROM users WHERE email="%s"`, body.Email,
				)
				err = dbh.QueryRow(statement).Scan(&id)
			}

			// Add user ID to request params to be used by handler
			idString := strconv.Itoa(id)
			ps = append(ps, httprouter.Param{"id", idString})
		}

		statement := fmt.Sprintf(
			`SELECT COUNT(*) FROM users WHERE id=%d AND sha256="%s"`, id, hash,
		)
		var count int
		err = dbh.QueryRow(statement).Scan(&count)

		if err == nil && count > 0 {
			log.Printf("Authenticated user with ID %d\n", id)
			handler(w, r, ps)
		} else {
			if err != nil {
				log.Println(err.Error())
			}
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	}
}

func (server *Server) Initialise(dbh *sql.DB) {
	server.Router = httprouter.New()
	server.dbh = dbh

	server.Router.POST("/user", Auth(endpoints.LogInUser(dbh), dbh))
	server.Router.GET("/user/:id/outgoings", Auth(endpoints.GetUserOutgoings(dbh), dbh))
	server.Router.POST("/user/:id/add", Auth(endpoints.AddOutgoing(dbh), dbh))
}

func (server *Server) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, server.Router))
}
