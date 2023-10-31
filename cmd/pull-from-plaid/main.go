package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gcjensen/amex"
	"github.com/gcjensen/splend-api/config"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	apiTimeout = 5 * time.Second
)

type splendClient struct {
	Token string
	URL   string

	client http.Client
}

type PlaidTransactionsRequest struct {
	ClientID    string `json:"client_id"`
	Secret      string `json:"secret"`
	AccessToken string `json:"access_token"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

type PlaidTransactionsResponse struct {
	PlaidTransactions []PlaidTransactions `json:"transactions"`
}

type PlaidTransactions struct {
	Amount               float32 `json:"amount"`
	Name                 string  `json:"name"`
	TransactionID        string  `json:"transaction_id"`
	PendingTransactionID string  `json:"pending_transaction_id"`
}

func main() {
	userID := os.Getenv("ID")
	token := os.Getenv("TOKEN")
	clientID := os.Getenv("CLIENT_ID")
	secret := os.Getenv("SECRET")
	accessToken := os.Getenv("ACCESS_TOKEN")

	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -4).Format("2006-01-02")

	log.Println("Fetching transactions from plaid...")

	ctx := context.Background()
	req := &PlaidTransactionsRequest{
		ClientID:    clientID,
		Secret:      secret,
		AccessToken: accessToken,
		StartDate:   yesterday,
		EndDate:     today,
	}
	rsp, err := fetchTransactions(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%d pending transactions fetched\n", len(rsp.PlaidTransactions))

	httpClient := http.Client{Timeout: apiTimeout}
	config := config.Load()
	apiURL := fmt.Sprintf("https://%s:%d", config.Host, config.Port)

	client := splendClient{
		Token:  token,
		URL:    apiURL,
		client: httpClient,
	}

	log.Println("Posting transactions to splend")

	for _, tx := range rsp.PlaidTransactions {
		// We're not interested in negatives amounts i.e. credits
		if tx.Amount < 0 {
			log.Println("Transaction amount is negative, ignoring.")
			continue
		}

		// Favour the pending tx ID, so that we don't have dupes when the tx settles
		txID := tx.PendingTransactionID
		if txID == "" {
			txID = tx.TransactionID
		}

		client.postToSplend(ctx, userID, &amex.Transaction{
			Amount:      int(tx.Amount * 100),
			Description: tx.Name,
			ID:          txID[:32],
		})
	}

	log.Println("Done.")
}

func fetchTransactions(ctx context.Context, body *PlaidTransactionsRequest) (*PlaidTransactionsResponse, error) {
	bodyJson, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := "https://development.plaid.com/transactions/get"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyJson))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	client := http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	var plaidRsp *PlaidTransactionsResponse
	if err := json.Unmarshal(rspBody, &plaidRsp); err != nil {
		return nil, err
	}

	return plaidRsp, err
}

func (cl *splendClient) postToSplend(ctx context.Context, userID string, tx *amex.Transaction) {
	txJSON, err := json.Marshal(tx)
	if err != nil {
		log.Println(err)
		return
	}

	url := cl.URL + fmt.Sprintf("/user/%s/amex", userID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(txJSON))
	if err != nil {
		log.Println(err)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Token", cl.Token)

	resp, err := cl.client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)

		return
	}

	var response map[string]string
	if err := json.Unmarshal(body, &response); err != nil {
		log.Println(err)
		return
	}

	log.Println(response)
}
