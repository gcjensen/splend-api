package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gcjensen/splend-api/api"
	"github.com/gcjensen/splend-api/config"
	"github.com/gcjensen/splend-api/splend"
	"github.com/gcjensen/splend-api/test"
	"github.com/julienschmidt/httprouter"
)

func TestGetUserMonthBreakdown(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(test.RandomUser(), test.RandomSha256(), dbh)

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
	router.GET("/user/:id/outgoings/breakdown/:month", api.GetUserMonthBreakdown(dbh))

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

func TestGetUserOutgoingsEndPoint(t *testing.T) {
	dbh := config.TestDBH()

	user, _ := splend.NewUser(test.RandomUser(), test.RandomSha256(), dbh)
	_ = user.AddOutgoing(test.RandomOutgoing())
	outgoings, _ := user.GetOutgoings(nil)
	outgoing := outgoings[0]

	router := httprouter.New()
	router.GET("/user/:id/outgoings", api.GetUserOutgoings(dbh))

	url := fmt.Sprintf("/user/%d/outgoings?description=%s", *user.ID, outgoing.Description)
	req, _ := http.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	splitTime := strings.Split(outgoing.Timestamp.String(), " ")
	timestamp := splitTime[0] + "T" + splitTime[1] + "Z"

	expected := fmt.Sprintf(`[{`+
		`"id":%d,`+
		`"description":"`+outgoing.Description+`",`+
		`"amount":"%d",`+
		`"owed":"%d",`+
		`"spender":"%d",`+
		`"category":"`+outgoing.Category+`",`+
		`"settled":null,`+
		`"timestamp":"`+timestamp+`"`+
		`}]`, *outgoing.ID, outgoing.Amount, outgoing.Owed, *user.ID)

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestLogInUser(t *testing.T) {
	dbh := config.TestDBH()

	tempUser, _ := splend.NewUser(test.RandomUser(), test.RandomSha256(), dbh)

	randomUser := test.RandomUser()
	randomUser.CoupleID = tempUser.CoupleID
	testUser, _ := splend.NewUser(randomUser, test.RandomSha256(), dbh)

	router := httprouter.New()
	router.POST("/user/:id", api.LogInUser(dbh))

	id := strconv.Itoa(*testUser.ID)
	req, _ := http.NewRequest("POST", "/user/"+id, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := fmt.Sprintf(`{`+
		`"id":%d,`+
		`"firstName":"`+testUser.FirstName+`",`+
		`"lastName":"`+testUser.LastName+`",`+
		`"email":"`+testUser.Email+`",`+
		`"colour":"`+*testUser.Colour+`",`+
		`"partner":{`+
		`"id":%d,`+
		`"firstName":"`+testUser.Partner.FirstName+`",`+
		`"lastName":"`+testUser.Partner.LastName+`",`+
		`"email":"`+testUser.Partner.Email+`",`+
		`"colour":"`+*testUser.Partner.Colour+`",`+
		`"partner":null,`+
		`"iconLink":null},`+
		`"iconLink":null`+
		`}`, *testUser.ID, *testUser.Partner.ID)

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
