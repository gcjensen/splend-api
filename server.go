package main

import (
	"crypto/sha256"
	"database/sql"
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

		statement := fmt.Sprintf(
			`SELECT id FROM users WHERE id=%d AND sha256="%s"`, id, hash,
		)

		var userID int
		err = dbh.QueryRow(statement).Scan(&userID)

		if err == nil && userID == id {
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

	server.Router.GET("/user/:id/details", Auth(endpoints.GetUserDetails(dbh), dbh))
	server.Router.GET("/user/:id/outgoings", Auth(endpoints.GetUserOutgoings(dbh), dbh))
}

func (server *Server) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, server.Router))
}
