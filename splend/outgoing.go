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

	dbh *sql.DB `json:"-"`
}

func NewOutgoing(outgoing *Outgoing, dbh *sql.DB) (*Outgoing, error) {
	o := outgoing
	o.dbh = dbh
	err := o.getInsertDetails()

	return o, err
}

func NewOutgoingFromDB(id int, dbh *sql.DB) (*Outgoing, error) {
	o := &Outgoing{ID: &id}
	o.dbh = dbh
	err := o.getOutgoing()

	return o, err
}

func (o *Outgoing) Delete() error {
	statement, err := o.dbh.Prepare(`
		DELETE FROM amex_transactions WHERE outgoing_id = ?
	`)
	if err != nil {
		return nil
	}

	defer statement.Close()

	_, err = statement.Exec(*o.ID)

	if err != nil {
		return nil
	}

	statement, _ = o.dbh.Prepare(`DELETE FROM outgoings WHERE id = ?`)
	defer statement.Close()

	_, err = statement.Exec(*o.ID)

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
	defer statement.Close()

	_, err = statement.Exec(
		o.Description, o.Amount, o.Owed, categoryID, *o.ID,
	)

	return err
}

func (o *Outgoing) getInsertDetails() error {
	err := o.getOutgoing()

	if errors.Is(err, ErrOutgoingUnknown) {
		err = o.insertOutgoing()
	}

	return err
}

func (o *Outgoing) getOutgoing() error {
	if o.ID == nil {
		return ErrOutgoingUnknown
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

	if errors.Is(err, sql.ErrNoRows) {
		return ErrOutgoingUnknown
	}

	return err
}

func (o *Outgoing) insertCategory() (*int, error) {
	statement, _ := o.dbh.Prepare(
		`INSERT INTO categories (name) VALUES (?)`,
	)
	defer statement.Close()

	_, err := statement.Exec(o.Category)
	if err != nil {
		return nil, err
	}

	var categoryID int
	err = o.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&categoryID)

	if err != nil {
		return nil, err
	}

	return &categoryID, nil
}

func (o *Outgoing) insertOutgoing() error {
	var categoryID *int

	err := o.dbh.QueryRow(
		`SELECT id FROM categories WHERE name = ?`, o.Category,
	).Scan(&categoryID)

	if errors.Is(err, sql.ErrNoRows) {
		categoryID, err = o.insertCategory()
	}

	if err != nil {
		return nil
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
	defer statement.Close()

	_, err = statement.Exec(o.Description, o.Amount, o.Owed, o.Spender, &categoryID)
	if err != nil {
		return err
	}

	return o.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&o.ID)
}
