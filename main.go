package main

import (
	"fmt"
	"github.com/gcjensen/splend-api/config"
	"log"
)

func main() {
	dbh := config.SplendDBH()
	config := config.Load()

	server := Server{}
	server.Initialise(dbh)

	log.Printf("API available at '%s:%d'", config.Host, config.Port)

	server.Run(fmt.Sprintf("%s:%d", config.Host, config.Port))
}
