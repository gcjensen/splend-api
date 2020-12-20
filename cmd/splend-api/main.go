package main

import (
	"fmt"
	"log"

	"github.com/gcjensen/splend-api/api"
	"github.com/gcjensen/splend-api/config"
)

func main() {
	dbh := config.SplendDBH()
	config := config.Load()

	server := api.NewServer()
	server.Initialise(dbh)

	log.Printf("API available at '%s:%d'", config.Host, config.Port)

	server.Run(fmt.Sprintf("%s:%d", config.Host, config.Port))
}
