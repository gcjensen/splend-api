package config

import (
	"database/sql"
	"fmt"
	"time"
)

func InsertTestUser(
	firstName string,
	lastName string,
	email string,
	coupleID int,
	dbh *sql.DB,
) int {
	statement := fmt.Sprintf(`
		INSERT INTO users
		(first_name, last_name, email, couple_id)
		VALUES ("%s", "%s", "%s", %d)`,
		firstName, lastName, email, coupleID)

	dbh.Exec(statement)

	var id int
	dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&id)
	return id
}

func InsertTestCouple(dbh *sql.DB) int {
	statement := fmt.Sprintf(
		`INSERT INTO couples (joining_date) VALUES ("2018-01-01")`,
	)

	dbh.Exec(statement)

	var id int
	dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&id)
	return id
}

func InsertTestOutgoing(
	description string,
	amount float64,
	owed float64,
	spender int,
	timestamp time.Time,
	dbh *sql.DB,
) int {
	statement := `INSERT INTO categories (id, name) VALUES (1, "General")`
	dbh.Exec(statement)

	statement = fmt.Sprintf(`
		INSERT INTO outgoings
		(description, amount, owed, spender_id, category_id, settled, timestamp)
		VALUES ("%s", %f, %f, %d, %d, NULL, "%s")`,
		description, amount, owed, spender, 1, timestamp)

	dbh.Exec(statement)

	var id int
	dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&id)
	return id
}

func DeleteAllData(dbh *sql.DB) {
	dbh.Exec("DELETE FROM outgoings")
	dbh.Exec("ALTER TABLE outgoings AUTO_INCREMENT = 1")
	dbh.Exec("DELETE FROM users")
	dbh.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
	dbh.Exec("DELETE FROM couples")
	dbh.Exec("ALTER TABLE couples AUTO_INCREMENT = 1")
}
