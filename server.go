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

		statement := fmt.Sprintf(
			`SELECT id FROM users WHERE sha256="%s"`, hash,
		)
		var id int
		err := dbh.QueryRow(statement).Scan(&id)

		// Add user ID to request params to be used by handler
		idString := strconv.Itoa(id)
		ps = append(ps, httprouter.Param{"id", idString})

		if err == nil {
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
	server.Router.POST(
		"/outgoing/settle/:outgoingID/:shouldSettle",
		Auth(endpoints.SettleOutgoing(dbh), dbh),
	)
	server.Router.POST(
		"/outgoing/delete/:outgoingID",
		Auth(endpoints.DeleteOutgoing(dbh), dbh),
	)
}

func (server *Server) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, server.Router))
}
