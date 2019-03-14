package config

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type config struct {
	Database struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"database"`

	TestDatabase struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"test_database"`

	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func SplendDBH() *sql.DB {
	c := Load()
	connectionString := fmt.Sprintf(
		"%s:%s@/%s?%s",
		c.Database.Username,
		c.Database.Password,
		c.Database.Name,
		"parseTime=true",
	)

	dbh, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	return dbh
}

func TestDBH() *sql.DB {
	c := Load()
	connectionString := fmt.Sprintf(
		"%s:%s@/%s?%s",
		c.TestDatabase.Username,
		c.TestDatabase.Password,
		c.TestDatabase.Name,
		"parseTime=true",
	)

	dbh, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	return dbh
}

func Load() *config {

	// Update to point towards your config file
	configFile, err := ioutil.ReadFile("/etc/splend.yaml")

	if err != nil {
		log.Printf("configFile.Get err #%v", err)
	}
	var c *config
	err = yaml.Unmarshal(configFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}
