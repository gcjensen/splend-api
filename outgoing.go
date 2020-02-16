package splend

import (
	"database/sql"
	"errors"
	"fmt"
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

func (o *Outgoing) Delete() error {
	statement, _ := o.dbh.Prepare(`DELETE FROM outgoings WHERE id = ?`)
	_, err := statement.Exec(*o.ID)

	return err
}

func (o *Outgoing) ToggleSettled(shouldSettle bool) error {
	var statement *sql.Stmt
	if shouldSettle {
		statement, _ = o.dbh.Prepare(`UPDATE outgoings SET settled = NOW() WHERE id= ?`)
	} else {
		statement, _ = o.dbh.Prepare(`UPDATE outgoings SET settled = NULL WHERE id= ?`)
	}

	_, err := statement.Exec(*o.ID)
	if err != nil {
		return err
	}

	err = o.getInsertDetails()

	return err
}

func (o *Outgoing) Update() error {
	var categoryID int

	err := o.dbh.QueryRow(
		`SELECT id FROM categories WHERE name = ?`, o.Category,
	).Scan(&categoryID)
	if err != nil {
		return err
	}

	statement, _ := o.dbh.Prepare(`
		UPDATE outgoings
		SET description = ?, amount = ?, owed = ?, category_id = ?
		WHERE id = ?
	`)

	_, err = statement.Exec(
		o.Description, o.Amount, o.Owed, categoryID, *o.ID,
	)

	return err
}

/************************** Private Implementation ****************************/

func (o *Outgoing) getInsertDetails() error {
	err := o.getOutgoing()
	if err != nil {
		var categoryID int

		err := o.dbh.QueryRow(
			`SELECT id FROM categories WHERE name = ?`, o.Category,
		).Scan(&categoryID)

		if err != nil {
			statement, _ := o.dbh.Prepare(
				`INSERT INTO categories (name) VALUES (?)`,
			)

			_, err = statement.Exec(o.Category)
			if err != nil {
				return err
			}

			err := o.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&categoryID)
			if err != nil {
				return err
			}
		}

		var settled string
		if o.Owed == 0 {
			settled = "NOW()"
		} else {
			settled = "NULL"
		}

		// #nosec - 'settled' is set above, so there's no risk of sql injection
		statement, _ := o.dbh.Prepare(fmt.Sprintf(`
			INSERT INTO outgoings
			(description, amount, owed, spender_id, category_id, settled,
			timestamp)
			VALUES (?, ?, ?, ?, ?, %s, NOW())
		`, settled))

		_, err = statement.Exec(o.Description, o.Amount, o.Owed, o.Spender, categoryID)
		if err != nil {
			return err
		}

		err = o.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&o.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *Outgoing) getOutgoing() error {
	if o.ID == nil {
		return errors.New("unknown outgoing")
	}

	query := `
		SELECT description, amount, owed, spender_id, c.name, settled, timestamp
		FROM outgoings o JOIN categories c ON o.category_id=c.id
		WHERE o.id=?
	`
	err := o.dbh.QueryRow(query, *o.ID).Scan(
		&o.Description,
		&o.Amount,
		&o.Owed,
		&o.Spender,
		&o.Category,
		&o.Settled,
		&o.Timestamp,
	)

	if err != nil {
		return errors.New("unknown outgoing")
	}

	return err
}
