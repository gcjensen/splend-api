package http

import (
	"fmt"
	"github.com/gcjensen/splend-api"
	"github.com/gcjensen/splend-api/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetUserMonthBreakdown(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(randomUser(), randomSha256(), dbh)

	groceries := &splend.Outgoing{
		Description: "Weekly shop",
		Amount:      5000,
		Owed:        2500,
		Category:    "Groceries",
	}

	groceriesAgain := &splend.Outgoing{
		Description: "Weekly shop",
		Amount:      4000,
		Owed:        2000,
		Category:    "Groceries",
	}

	beers := &splend.Outgoing{
		Description: "Beers",
		Amount:      2000,
		Owed:        0,
		Category:    "Drinks",
	}
	_ = user.AddOutgoing(groceries)
	_ = user.AddOutgoing(groceriesAgain)
	_ = user.AddOutgoing(beers)

	router := httprouter.New()
	router.GET("/user/:id/outgoings/breakdown/:month", GetUserMonthBreakdown(dbh))

	month := time.Now().Format("2006-01")
	url := fmt.Sprintf("/user/%d/outgoings/breakdown/%s", *user.ID, month)
	req, _ := http.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `[` +
		`{"category":"Groceries","user_total":4500,"couple_total":9000},` +
		`{"category":"Drinks","user_total":2000,"couple_total":0}` +
		`]`

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
