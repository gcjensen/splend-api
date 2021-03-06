package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gcjensen/amex"
	"github.com/gcjensen/splend-api/config"
)

const (
	amexTimeout = 60 * time.Second
	apiTimeout  = 5 * time.Second
)

type splendClient struct {
	Token string
	URL   string

	client http.Client
	wg     sync.WaitGroup
}

func main() {
	userID := os.Getenv("ID")
	amexUserID := os.Getenv("USER_ID")
	amexPassword := os.Getenv("PASSWORD")
	token := os.Getenv("TOKEN")

	log.Println("Fetching pending transactions from Amex")

	ctx, cancel := context.WithTimeout(context.Background(), amexTimeout)
	defer cancel()

	a, _ := amex.NewContext(ctx, amexUserID, amexPassword)

	transactions, err := a.GetPendingTransactions()
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("%d pending transactions fetched", len(transactions))

	a.Close()

	httpClient := http.Client{Timeout: apiTimeout}
	config := config.Load()
	apiURL := fmt.Sprintf("http://%s:%d", config.Host, config.Port)

	client := splendClient{
		Token:  token,
		URL:    apiURL,
		client: httpClient,
		wg:     sync.WaitGroup{},
	}

	log.Println("Posting transactions to splend")

	for _, tx := range transactions {
		// We're not interested in negatives amounts i.e. credits
		if tx.Amount < 0 {
			log.Println("Transaction amount is negative, ignoring.")
			continue
		}

		inc := 1 // Linter doesn't allow magic numbers
		client.wg.Add(inc)

		go client.postToSplend(ctx, userID, tx)
	}

	client.wg.Wait()
	log.Println("Done.")
}

func (cl *splendClient) postToSplend(ctx context.Context, userID string, tx *amex.Transaction) {
	defer cl.wg.Done()

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
