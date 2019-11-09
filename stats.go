package splend

import (
	"database/sql"
	"fmt"
)

type CategoryTotal struct {
	Category string `json:"category"`
	Total    int    `json:"total"`
}

func (self *User) GetMonthBreakdown(
	monthYear string,
	forCouple bool,
) ([]CategoryTotal, error) {

	var partnerID int
	if self.Partner.ID != nil {
		partnerID = *self.Partner.ID
	}

	query := `
		SELECT c.name, SUM(%s)
		FROM outgoings o
		JOIN categories c ON o.category_id=c.id
		WHERE spender_id in (?, ?)
		AND DATE_FORMAT(timestamp,'%%Y-%%m') = ?
		group by 1
		order by 2 desc;
	`

	var rows *sql.Rows
	var err error
	if !forCouple {
		rows, err = self.dbh.Query(
			fmt.Sprintf(query, "IF(spender_id = ?, amount - owed, owed)"),
			self.ID,
			self.ID,
			partnerID,
			monthYear,
		)
	} else {
		rows, err = self.dbh.Query(
			fmt.Sprintf(query, "amount"),
			self.ID,
			partnerID,
			monthYear,
		)
	}

	defer rows.Close()

	if err != nil {
		return nil, err
	}

	var totals []CategoryTotal

	for rows.Next() {
		var c CategoryTotal
		if err := rows.Scan(&c.Category, &c.Total); err != nil {
			return nil, err
		}
		totals = append(totals, c)
	}

	return totals, err
}
