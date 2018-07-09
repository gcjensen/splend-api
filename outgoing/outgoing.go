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

func New(id int, dbh *sql.DB) (*Outgoing, error) {

	self := &Outgoing{ID: &id}
	self.dbh = dbh

	err := self.getInsertDetails(dbh)

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

	err = self.getInsertDetails(self.dbh)
	return err
}

/************************** Private Implementation ****************************/

func (self *Outgoing) getInsertDetails(dbh *sql.DB) error {
	err := self.getOutgoing(dbh)
	if err != nil {
		// TODO implement this and use it in User.AddOutgoing
		return errors.New("Proper outgoing creation not yet implemented")
	}

	return nil
}

func (self *Outgoing) getOutgoing(dbh *sql.DB) error {
	statement := fmt.Sprintf(`
		SELECT description, amount, owed, spender_id, c.name, settled, timestamp
		FROM outgoings o JOIN categories c ON o.category_id=c.id
		WHERE o.id="%d"`, *self.ID)

	err := dbh.QueryRow(statement).Scan(
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
