package api

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type Server struct {
	Router *httprouter.Router
	dbh    *sql.DB
}

func NewServer() Server {
	return Server{}
}

func Auth(handler httprouter.Handle, dbh *sql.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		token := r.Header.Get("Token")
		// Sha256 the token and check against what we have stored for the user
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))

		statement := `SELECT id FROM users WHERE sha256=?`

		var id int

		err := dbh.QueryRow(statement, hash).Scan(&id)

		// Add user ID to request params to be used by handler
		idString := strconv.Itoa(id)

		ps = append(ps, httprouter.Param{Key: "id", Value: idString})

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

	server.Router.POST("/user", Auth(LogInUser(dbh), dbh))
	server.Router.GET("/user/:id/outgoings", Auth(GetUserOutgoings(dbh), dbh))
	server.Router.GET("/user/:id/summary", Auth(GetUserSummary(dbh), dbh))
	server.Router.GET("/user/:id/breakdown/:month", Auth(GetUserMonthBreakdown(dbh), dbh))
	server.Router.POST("/user/:id/add", Auth(AddOutgoing(dbh), dbh))
	server.Router.POST("/user/:id/amex", Auth(AddFromAmex(dbh), dbh))
	server.Router.POST("/user/:id/monzo-webhook", AddFromMonzo(dbh))
	server.Router.POST("/outgoing/settle/:outgoingID/:shouldSettle", Auth(SettleOutgoing(dbh), dbh))
	server.Router.POST("/outgoing/delete/:outgoingID", Auth(DeleteOutgoing(dbh), dbh))
	server.Router.POST("/outgoing/update/:outgoingID", Auth(UpdateOutgoing(dbh), dbh))
}

func (server *Server) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, server.Router))
}
