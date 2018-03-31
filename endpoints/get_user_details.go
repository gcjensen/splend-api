package endpoints

import (
	"database/sql"
	"fmt"
	"github.com/gcjensen/settle-api/user"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

func GetUserDetails(dbh *sql.DB) httprouter.Handle {
	return httprouter.Handle(func(
		writer http.ResponseWriter,
		req *http.Request,
		params httprouter.Params,
	) {

		// Pull out into some sort of reuable param verification logic
		id, err := strconv.Atoi(params.ByName("id"))
		if err != nil {
			fmt.Println(err)
			respondWithError(err, writer)
			return
		}

		user, err := user.NewFromDB(id, dbh)
		if err != nil {
			fmt.Println(err)
			respondWithError(err, writer)
			return
		}

		respondWithJSON(writer, http.StatusOK, user)
	})
}
