package splend

import (
	"database/sql"
	"errors"
	"time"
)

type Outgoing struct {
	ID          *int       `json:"id"`
	Description string     `json:"description"`
	Amount      int        `json:"amount,string"`
	Owed        int        `json:"owed,string"`
	Spender     int        `json:"spender,string"`
	Category    string     `json:"category"`
	Settled     *time.Time `json:"settled"`
	Timestamp   *time.Time `json:"timestamp"`
	dbh         *sql.DB
}

func NewOutgoing(outgoing *Outgoing, dbh *sql.DB) (*Outgoing, error) {
	self := outgoing
	self.dbh = dbh
	err := self.getInsertDetails()
	self.dbh = nil
	return self, err
}

func NewOutgoingFromDB(id int, dbh *sql.DB) (*Outgoing, error) {
	self := &Outgoing{ID: &id}
	self.dbh = dbh
	err := self.getOutgoing()
	return self, err
}

func (self *Outgoing) Delete() error {
	statement, _ := self.dbh.Prepare(`DELETE FROM outgoings WHERE id = ?`)
	_, err := statement.Exec(*self.ID)

	return err
}

func (self *Outgoing) ToggleSettled(shouldSettle bool) error {
	var settled sql.NullString
	if shouldSettle {
		settled = sql.NullString{"NOW()", true}
	} else {
		settled = sql.NullString{}
	}

	statement, _ := self.dbh.Prepare(`
		UPDATE outgoings SET settled = ? WHERE id= ?
	`)

	_, err := statement.Exec(settled, *self.ID)
	if err != nil {
		return err
	}

	err = self.getInsertDetails()
	return err
}

func (self *Outgoing) Update() error {
	var categoryID int
	err := self.dbh.QueryRow(
		`SELECT id FROM categories WHERE name = ?`, self.Category,
	).Scan(&categoryID)

	statement, _ := self.dbh.Prepare(`
		UPDATE outgoings
		SET description = ?, amount = ?, owed = ?, category_id = ?
		WHERE id = ?
	`)

	_, err = statement.Exec(
		self.Description, self.Amount, self.Owed, categoryID, *self.ID,
	)

	return err
}

/************************** Private Implementation ****************************/

func (self *Outgoing) getInsertDetails() error {
	err := self.getOutgoing()
	if err != nil {

		var categoryID int
		err := self.dbh.QueryRow(
			`SELECT id FROM categories WHERE name = ?`, self.Category,
		).Scan(&categoryID)

		if err != nil {
			statement, _ := self.dbh.Prepare(
				`INSERT INTO categories (name) VALUES (?)`,
			)
			_, err = statement.Exec(self.Category)
			if err != nil {
				return err
			}

			self.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&categoryID)
		}

		var settled sql.NullString
		if self.Owed == 0 {
			settled = sql.NullString{"NOW()", true}
		} else {
			settled = sql.NullString{}
		}

		statement, _ := self.dbh.Prepare(`
			INSERT INTO outgoings
			(description, amount, owed, spender_id, category_id, settled,
			timestamp)
			VALUES (?, ?, ?, ?, ?, ?, NOW())
		`)

		_, err = statement.Exec(
			self.Description, self.Amount, self.Owed, self.Spender, categoryID,
			settled,
		)

		self.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&self.ID)
	}

	return nil
}

func (self *Outgoing) getOutgoing() error {
	if self.ID == nil {
		return errors.New("Unknown outgoing")
	}

	query := `
		SELECT description, amount, owed, spender_id, c.name, settled, timestamp
		FROM outgoings o JOIN categories c ON o.category_id=c.id
		WHERE o.id=?
	`
	err := self.dbh.QueryRow(query, *self.ID).Scan(
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
