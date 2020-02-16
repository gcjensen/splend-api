package splend

import (
	"database/sql"
	"errors"
)

type LinkedAccounts struct {
	Monzo *string `json:"-"`
}

type User struct {
	dbh            *sql.DB
	ID             *int    `json:"id"`
	FirstName      string  `json:"firstName"`
	LastName       string  `json:"lastName"`
	Email          string  `json:"email"`
	Colour         *string `json:"colour"`
	Partner        *User   `json:"partner"`
	CoupleID       *int    `json:"-"`
	IconLink       *string `json:"iconLink"`
	LinkedAccounts `json:"-"`
}

func NewUser(user *User, sha256 string, dbh *sql.DB) (*User, error) {
	self := user
	self.dbh = dbh

	err := self.getInsertDetails(sha256)

	return self, err
}

func NewUserFromDB(id int, dbh *sql.DB) (*User, error) {
	self := &User{dbh: dbh, LinkedAccounts: LinkedAccounts{}}
	err := dbh.QueryRow(
		`SELECT email FROM users WHERE id = ?`, id,
	).Scan(&self.Email)

	if err != nil {
		return nil, errors.New("unknown user")
	}

	err = self.getUser()

	return self, err
}

func (u *User) AddOutgoing(o *Outgoing) error {
	o.Spender = *u.ID
	_, err := NewOutgoing(o, u.dbh)

	return err
}

func (u *User) GetOutgoings() ([]Outgoing, error) {
	var partnerID int
	if u.Partner.ID != nil {
		partnerID = *u.Partner.ID
	}

	query := `
		SELECT o.id, description, amount, owed, spender_id, c.name, settled,
		timestamp
		FROM outgoings o
		JOIN categories c ON o.category_id=c.id
		WHERE spender_id = ? OR (spender_id = ? AND owed > 0)
		ORDER BY o.timestamp DESC
	`

	rows, err := u.dbh.Query(query, u.ID, partnerID)
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

func (u *User) LinkAccounts(accounts *LinkedAccounts) error {
	if u.LinkedAccounts.Monzo != nil {
		return errors.New("monzo account already linked")
	}

	u.LinkedAccounts.Monzo = accounts.Monzo
	statement, _ := u.dbh.Prepare(
		`INSERT INTO linked_accounts (user_id, monzo) VALUES (?, ?)`,
	)

	_, err := statement.Exec(*u.ID, *accounts.Monzo)

	return err
}

/************************** Private Implementation ****************************/

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
	if err != nil {
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

		_, err = statement.Exec(u.FirstName, u.LastName, u.Email, *u.CoupleID, *u.Colour, sha256)
		if err != nil {
			return err
		}

		err = u.dbh.QueryRow("SELECT LAST_INSERT_ID()").Scan(&u.ID)
		if err != nil {
			return err
		}

		err = u.getPartner()
		if err != nil {
			return err
		}
	}

	return nil
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
		return errors.New("unknown user")
	}

	err = u.getPartner()
	if err != nil {
		return err
	}

	statement := "SELECT monzo FROM linked_accounts WHERE user_id= ?"
	_ = u.dbh.QueryRow(statement, *u.ID).Scan(&u.LinkedAccounts.Monzo)

	return nil
}

func (u *User) getPartner() error {
	if u.CoupleID == nil {
		return errors.New("no partner")
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

	if err != sql.ErrNoRows {
		return err
	}

	return nil
}
