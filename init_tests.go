package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"strings"
)

const FilePath = "meta/schema.sql"

var db *sql.DB

func init() {
	connectionString := fmt.Sprintf(
		"%s:%s@/%s?%s",
		"test",
		"scrabble",
		"settle_test",
		"parseTime=true")

	var err error
	db, err = sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}
}

func ResetDB() *sql.DB {
	file, err := ioutil.ReadFile(FilePath)

	if err != nil {
		fmt.Println(err)
	}

	requests := strings.Split(string(file), ";")

	for _, request := range requests {
		db.Exec(request)
	}

	return db
}
