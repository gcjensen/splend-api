package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	Router      *mux.Router
	dbh         *sql.DB
	userHandler *UserHandler
}

type authMiddleware struct {
	dbh *sql.DB
}

func (amw *authMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token := r.Header.Get("Token")
		// Sha256 the token and check against what we have stored for the user
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
		params := mux.Vars(r)
		id, err := strconv.Atoi(params["id"])

		statement := fmt.Sprintf(
			`SELECT id FROM users WHERE id=%d AND sha256="%s"`, id, hash,
		)

		var userID int
		err = amw.dbh.QueryRow(statement).Scan(&userID)

		if err == nil && userID == id {
			log.Printf("Authenticated user with ID %d\n", id)
			next.ServeHTTP(w, r)
		} else {
			if err != nil {
				log.Println(err.Error())
			}
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}

func (server *Server) Initialise(dbh *sql.DB) {
	server.Router = mux.NewRouter()
	server.dbh = dbh
	server.userHandler = &UserHandler{server.dbh}

	amw := authMiddleware{server.dbh}
	server.Router.Use(amw.Middleware)

	// Route for retrieving all details on a user
	server.Router.HandleFunc(
		"/user/{id:[0-9]+}/details",
		server.userHandler.GetDetails,
	).Methods("GET")
}

func (server *Server) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, server.Router))
}
