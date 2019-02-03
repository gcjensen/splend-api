package outgoing

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Outgoing struct {
	ID          *int       `json:"id"`
	Description string     `json:"description"`
	Amount      float64    `json:"amount,string"`
	Owed        float64    `json:"owed,string"`
	Spender     int        `json:"spender,string"`
	Category    string     `json:"category"`
	Settled     *time.Time `json:"settled"`
	Timestamp   *time.Time `json:"timestamp"`
	dbh         *sql.DB
}

func New(outgoing *Outgoing, dbh *sql.DB) (*Outgoing, error) {
	self := outgoing
	self.dbh = dbh
	err := self.getInsertDetails()
	self.dbh = nil
	return self, err
}

func NewFromDB(id int, dbh *sql.DB) (*Outgoing, error) {
	self := &Outgoing{ID: &id}
	self.dbh = dbh
	err := self.getOutgoing()
	return self, err
}

func (self *Outgoing) Delete() error {
	statement := fmt.Sprintf(`DELETE FROM outgoings WHERE id = %d`, *self.ID)
	_, err := self.dbh.Exec(statement)

	return err
}

func (self *Outgoing) ToggleSettled(settled bool) error {
	var statement string
	if settled {
		statement = fmt.Sprintf(
			`UPDATE outgoings SET settled = NOW() WHERE id=%d`, *self.ID,
		)
	} else {
		statement = fmt.Sprintf(
			`UPDATE outgoings SET settled = NULL WHERE id=%d`, *self.ID,
		)
	}

	_, err := self.dbh.Exec(statement)
	if err != nil {
		return err
	}

	err = self.getInsertDetails()
	return err
}

func (self *Outgoing) Update() error {
	statement := fmt.Sprintf(
		`SELECT id FROM categories WHERE name = "%s"`, self.Category,
	)
	var categoryID int
	err := self.dbh.QueryRow(statement).Scan(&categoryID)

	statement = fmt.Sprintf(`
		UPDATE outgoings
		SET description = "%s", amount = %f, owed = %f, category_id = %d
		WHERE id = %d`,
		self.Description, self.Amount, self.Owed, categoryID, *self.ID,
	)

	_, err = self.dbh.Exec(statement)

	return err
}

/************************** Private Implementation ****************************/

func (self *Outgoing) getInsertDetails() error {
	err := self.getOutgoing()
	if err != nil {
		statement := fmt.Sprintf(
			`SELECT id FROM categories WHERE name = "%s"`, self.Category,
		)

		var categoryID int
		err := self.dbh.QueryRow(statement).Scan(&categoryID)

		if err != nil {
			statement := fmt.Sprintf(`
				INSERT INTO categories (name) VALUES ("%s")`, self.Category,
			)
			_, err = self.dbh.Exec(statement)
			if err != nil {
				return err
			}

			self.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&categoryID)
		}

		statement = fmt.Sprintf(`
			INSERT INTO outgoings
			(description, amount, owed, spender_id, category_id, settled,
			timestamp)
			VALUES ("%s", %f, %f, %d, %d, NULL, NOW())`,
			self.Description, self.Amount, self.Owed, self.Spender, categoryID,
		)

		_, err = self.dbh.Exec(statement)

		self.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&self.ID)
	}

	return nil
}

func (self *Outgoing) getOutgoing() error {
	if self.ID == nil {
		return errors.New("Unknown outgoing")
	}

	statement := fmt.Sprintf(`
		SELECT description, amount, owed, spender_id, c.name, settled, timestamp
		FROM outgoings o JOIN categories c ON o.category_id=c.id
		WHERE o.id="%d"`, *self.ID)

	err := self.dbh.QueryRow(statement).Scan(
		&self.Description,
		&self.Amount,
		&self.Owed,
		&self.Spender,
		&self.Category,
		&self.Settled,
		&self.Timestamp,
	)

	if err != nil {
		return errors.New("Unknown outgoing")
	}

	return err
}
