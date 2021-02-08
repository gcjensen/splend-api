package splend

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gcjensen/amex"
)

type MonzoAccount struct {
	ID *string `json:"-"`
}

type User struct {
	dbh          *sql.DB
	ID           *int         `json:"id"`
	FirstName    string       `json:"firstName"`
	LastName     string       `json:"lastName"`
	Email        string       `json:"email"`
	Colour       *string      `json:"colour"`
	Partner      *User        `json:"partner"`
	CoupleID     *int         `json:"-"`
	IconLink     *string      `json:"iconLink"`
	MonzoAccount MonzoAccount `json:"-"`
}

type Summary struct {
	Balance int `json:"balance"`
}

//nolint
var whereClauseMappings = map[string]string{
	"months":      "timestamp > NOW() - INTERVAL ? MONTH",
	"description": "description LIKE ?",
}

func NewUser(user *User, sha256 string, dbh *sql.DB) (*User, error) {
	self := user
	self.dbh = dbh

	err := self.getInsertDetails(sha256)

	return self, err
}

func NewUserFromDB(id int, dbh *sql.DB) (*User, error) {
	self := &User{dbh: dbh, MonzoAccount: MonzoAccount{}}

	err := dbh.QueryRow(
		`SELECT email FROM users WHERE id = ?`, id,
	).Scan(&self.Email)
	if err != nil {
		return nil, ErrUserUnknown
	}

	err = self.getUser()

	return self, err
}

func (u *User) AddAmexTransaction(tx amex.Transaction) error {
	err := u.isAmexTransactionNew(tx)
	if !errors.Is(err, sql.ErrNoRows) {
		return ErrAlreadyExists
	}

	err = u.AddOutgoing(&Outgoing{
		Amount:      tx.Amount,
		Category:    "Other",
		Description: tx.Description,
		Spender:     *u.ID,
	})
	if err != nil {
		return err
	}

	var outgoingID int
	err = u.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&outgoingID)

	if err != nil {
		return err
	}

	statement, _ := u.dbh.Prepare(
		`INSERT INTO amex_transactions (amex_id, outgoing_id) VALUES (?, ?)`,
	)
	defer statement.Close()

	_, err = statement.Exec(tx.ID, outgoingID)

	return err
}

func (u *User) AddOutgoing(o *Outgoing) error {
	o.Spender = *u.ID
	_, err := NewOutgoing(o, u.dbh)

	return err
}

func (u *User) GetOutgoings(where map[string]interface{}) ([]Outgoing, error) {
	var partnerID int
	if u.Partner.ID != nil {
		partnerID = *u.Partner.ID
	}

	query := `
		SELECT o.id, description, amount, owed, spender_id, c.name, settled,
		timestamp
		FROM outgoings o
		JOIN categories c ON o.category_id=c.id
		WHERE (spender_id = ? OR (spender_id = ? AND owed > 0))
	`

	params := []interface{}{u.ID, partnerID}

	for field, value := range where {
		query += fmt.Sprintf("AND %s ", whereClauseMappings[field])

		params = append(params, value)
	}

	query += `ORDER BY o.timestamp DESC`

	rows, err := u.dbh.Query(query, params...)
	if err != nil {
		return nil, err
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	defer rows.Close()

	var outgoings []Outgoing

	for rows.Next() {
		var o Outgoing
		if err := rows.Scan(&o.ID, &o.Description, &o.Amount, &o.Owed,
			&o.Spender, &o.Category, &o.Settled, &o.Timestamp); err != nil {
			return nil, err
		}

		outgoings = append(outgoings, o)
	}

	return outgoings, err
}

func (u *User) GetSummary() (*Summary, error) {
	var partnerID int
	if u.Partner.ID != nil {
		partnerID = *u.Partner.ID
	}

	query := `
		SELECT SUM(IF(spender_id= ?, owed, 0) - IF(spender_id= ?, owed, 0))
		FROM outgoings WHERE settled is null;
	`

	var s Summary

	err := u.dbh.QueryRow(query, u.ID, partnerID).Scan(&s.Balance)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (u *User) LinkMonzo(account *MonzoAccount) error {
	if u.MonzoAccount.ID != nil {
		return ErrMonzoAccountAlreadyLinked
	}

	u.MonzoAccount.ID = account.ID

	statement, _ := u.dbh.Prepare(
		`INSERT INTO monzo_accounts (user_id, account_id) VALUES (?, ?)`,
	)
	defer statement.Close()

	_, err := statement.Exec(*u.ID, *account.ID)

	return err
}

func (u *User) addCouple() (*int, error) {
	_, err := u.dbh.Exec(`
		INSERT INTO couples (joining_date) VALUES (NOW())
	`)
	if err != nil {
		return nil, err
	}

	var id int

	err = u.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&id)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func (u *User) getInsertDetails(sha256 string) error {
	err := u.getUser()
	if errors.Is(err, ErrUserUnknown) {
		err = u.insertDetails(sha256)
	}

	return err
}

func (u *User) getUser() error {
	query := `
        SELECT id, first_name, last_name, couple_id, colour, icon_link
        FROM users
        WHERE email= ?
	`

	err := u.dbh.QueryRow(query, u.Email).Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.CoupleID,
		&u.Colour,
		&u.IconLink,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserUnknown
		}

		return err
	}

	err = u.getPartner()
	if err != nil {
		return err
	}

	statement := "SELECT account_id FROM monzo_accounts WHERE user_id= ?"
	_ = u.dbh.QueryRow(statement, *u.ID).Scan(&u.MonzoAccount.ID)

	return nil
}

func (u *User) getPartner() error {
	if u.CoupleID == nil {
		return ErrUserNotInCouple
	}

	query := `
        SELECT id, first_name, last_name, email, colour, couple_id, icon_link
		FROM users
		WHERE couple_id = ? AND id != ?
	`

	partner := &User{}
	err := u.dbh.QueryRow(query, u.CoupleID, u.ID).Scan(
		&partner.ID,
		&partner.FirstName,
		&partner.LastName,
		&partner.Email,
		&partner.Colour,
		&partner.CoupleID,
		&partner.IconLink,
	)
	u.Partner = partner

	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	return nil
}

func (u *User) insertDetails(sha256 string) error {
	if u.CoupleID == nil {
		coupleID, err := u.addCouple()
		if err != nil {
			return err
		}

		u.CoupleID = coupleID
	}

	statement, _ := u.dbh.Prepare(`
			INSERT INTO users
			(first_name, last_name, email, couple_id, colour, sha256)
			VALUES (?, ?, ?, ?, ?, ?)
		`)
	defer statement.Close()

	_, err := statement.Exec(u.FirstName, u.LastName, u.Email, *u.CoupleID, *u.Colour, sha256)
	if err != nil {
		return err
	}

	err = u.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&u.ID)
	if err != nil {
		return err
	}

	return u.getPartner()
}

func (u *User) isAmexTransactionNew(tx amex.Transaction) error {
	query := `
		SELECT * FROM amex_transactions a
		JOIN outgoings o ON a.outgoing_id=o.id
		WHERE a.amex_id = ? AND o.spender_id = ?
	`

	err := u.dbh.QueryRow(query, tx.ID, u.ID).Scan()

	return err
}
