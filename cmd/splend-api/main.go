package main

import (
	"fmt"
	"github.com/gcjensen/splend-api/config"
	"github.com/gcjensen/splend-api/http"
	"log"
)

func main() {
	dbh := config.SplendDBH()
	config := config.Load()

	server := http.NewServer()
	server.Initialise(dbh)

	log.Printf("API available at '%s:%d'", config.Host, config.Port)

	server.Run(fmt.Sprintf("%s:%d", config.Host, config.Port))
}
