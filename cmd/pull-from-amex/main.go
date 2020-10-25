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

func main() {
	userID := os.Getenv("ID")
	amexUserID := os.Getenv("USER_ID")
	amexPassword := os.Getenv("PASSWORD")
	token := os.Getenv("TOKEN")

	log.Println("Fetching pending transactions from Amex")

	ctx, cancel := context.WithTimeout(context.Background(), amexTimeout)

	a, _ := amex.NewContext(ctx, amexUserID, amexPassword)

	transactions, err := a.GetPendingTransactions()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%d pending transactions fetched", len(transactions))

	cancel()
	a.Close()

	httpClient := http.Client{Timeout: apiTimeout}
	config := config.Load()
	apiURL := fmt.Sprintf("http://%s:%d/user/%s/amex", config.Host, config.Port, userID)

	log.Println("Posting transactions to splend")

	var wg sync.WaitGroup

	for _, tx := range transactions {
		// We're not interested in negatives amounts i.e. credits
		if tx.Amount < 0 {
			log.Println("Transaction amount is negative, ignoring.")
			continue
		}

		inc := 1 // Linter doesn't allow magic numbers
		wg.Add(inc)

		go postToSplend(&wg, httpClient, apiURL, token, tx)
	}

	wg.Wait()
	log.Println("Done.")
}

func postToSplend(wg *sync.WaitGroup, cl http.Client, url, token string, tx *amex.Transaction) {
	defer wg.Done()

	txJSON, err := json.Marshal(tx)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(txJSON))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Token", token)

	resp, err := cl.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return
	}

	var response map[string]string
	if err := json.Unmarshal(body, &response); err != nil {
		log.Fatal(err)
		return
	}

	log.Println(response)
}
