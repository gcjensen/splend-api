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
	Tags        []string   `json:"tags"`
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
	o := &Outgoing{ID: &id, Tags: []string{}}
	o.dbh = dbh
	err := o.getOutgoing()

	return o, err
}

func (o *Outgoing) Delete() error {
	table := []string{
		"amex_transactions",
		"outgoing_tags",
	}

	for _, table := range table {
		statement, _ := o.dbh.Prepare(fmt.Sprintf(`DELETE FROM %s WHERE outgoing_id = ?`, table))
		if _, err := statement.Exec(*o.ID); err != nil {
			return nil
		}
		statement.Close()
	}

	statement, _ := o.dbh.Prepare(`DELETE FROM outgoings WHERE id = ?`)
	if _, err := statement.Exec(*o.ID); err != nil {
		return nil
	}
	statement.Close()

	return nil
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

	if err := o.dbh.QueryRow(`SELECT id FROM categories WHERE name = ?`, o.Category).Scan(&categoryID); err != nil {
		return err
	}

	statement, _ := o.dbh.Prepare(`
		UPDATE outgoings
		SET description = ?, amount = ?, owed = ?, category_id = ?
		WHERE id = ?
	`)
	defer statement.Close()

	if _, err := statement.Exec(o.Description, o.Amount, o.Owed, categoryID, *o.ID); err != nil {
		return err
	}

	return o.insertOutgoingTags()
}

func (o *Outgoing) UpdateTags(tags []string) error {
	tagSet := make(map[string]struct{})
	for _, tag := range tags {
		tagSet[tag] = struct{}{}
	}

	// Loop over the existing tags and delete any that aren't in the "new" list
	for _, tag := range o.Tags {
		if _, ok := tagSet[tag]; !ok {
			statement, _ := o.dbh.Prepare(`DELETE ot FROM outgoing_tags ot JOIN tags t ON ot.tag_id=t.id WHERE t.tag = ?`)
			if _, err := statement.Exec(tag); err != nil {
				return err
			}
			statement.Close()
		}
	}

	// Finally update the tags, inserting any that are new
	o.Tags = tags
	return o.insertOutgoingTags()
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

	if err := o.addTags(); err != nil {
		return err
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

	err := o.dbh.QueryRow(`SELECT id FROM categories WHERE name = ?`, o.Category).Scan(&categoryID)

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

	if o.Timestamp == nil {
		now := time.Now()
		o.Timestamp = &now
	}

	// #nosec - 'settled' is set above, so there's no risk of sql injection
	statement, _ := o.dbh.Prepare(fmt.Sprintf(`
			INSERT INTO outgoings
			(description, amount, owed, spender_id, category_id, settled,
			timestamp)
			VALUES (?, ?, ?, ?, ?, %s, ?)
		`, settled))
	defer statement.Close()

	_, err = statement.Exec(o.Description, o.Amount, o.Owed, o.Spender, &categoryID, o.Timestamp)
	if err != nil {
		return err
	}

	err = o.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&o.ID)
	if err != nil {
		return err
	}

	return o.insertOutgoingTags()
}

func (o *Outgoing) insertOutgoingTags() error {

	for _, tag := range o.Tags {
		var tagId *int
		err := o.dbh.QueryRow(`SELECT id FROM tags WHERE tag = ?`, tag).Scan(&tagId)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				tagId, err = o.insertTag(tag)
			} else {
				return err
			}
		}

		statement, _ := o.dbh.Prepare(`INSERT IGNORE INTO outgoing_tags (tag_id, outgoing_id) VALUES (?, ?)`)
		defer statement.Close()

		_, err = statement.Exec(&tagId, o.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *Outgoing) insertTag(tag string) (*int, error) {
	statement, _ := o.dbh.Prepare(
		`INSERT INTO tags (tag) VALUES (?)`,
	)
	defer statement.Close()

	_, err := statement.Exec(tag)
	if err != nil {
		return nil, err
	}

	var tagId int
	err = o.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&tagId)

	if err != nil {
		return nil, err
	}

	return &tagId, nil
}

func (o *Outgoing) addTags() error {
	query := `
		SELECT tag FROM tags t JOIN outgoing_tags ot ON t.id=ot.tag_id
		WHERE ot.outgoing_id=?
	`
	rows, err := o.dbh.Query(query, o.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return err
		}

		o.Tags = append(o.Tags, tag)
	}

	return nil
}
