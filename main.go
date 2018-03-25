package main

import (
	"github.com/Sirupsen/logrus"
)

func main() {
	a := Server{}
	a.Initialise("root", "your-root-password", "db-name")

	logger := logrus.New()
	logger.Infof("Server listening on port 3000")

	a.Run("0.0.0.0:3000")
}
