package main

import (
	"fmt"
	"github.com/gcjensen/settle-api/config"
	"log"
)

func main() {
	dbh := config.SettleDBH()
	config := config.Load()

	server := Server{}
	server.Initialise(dbh)

	log.Printf("API available at '%s:%d'", config.Host, config.Port)

	server.Run(fmt.Sprintf("%s:%d", config.Host, config.Port))
}
