package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

func main() {
	var c config
	c.getConfig()

	server := Server{}
	server.Initialise(c.Username, c.Password, c.Database)

	log.Printf("API available at '%s:%d'", c.Host, c.Port)

	server.Run(fmt.Sprintf("%s:%d", c.Host, c.Port))
}

type config struct {
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
}

func (c *config) getConfig() *config {

	// Update to point towards your config file
	configFile, err := ioutil.ReadFile("/etc/settle-api.yaml")

	if err != nil {
		log.Printf("configFile.Get err #%v", err)
	}
	err = yaml.Unmarshal(configFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}
