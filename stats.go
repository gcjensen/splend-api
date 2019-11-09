package splend

type CategoryTotals struct {
	Category    string `json:"category"`
	UserTotal   int    `json:"user_total"`
	CoupleTotal int    `json:"couple_total"`
}

func (self *User) GetMonthBreakdown(
	monthYear string,
) ([]CategoryTotals, error) {

	var partnerID int
	if self.Partner.ID != nil {
		partnerID = *self.Partner.ID
	}

	query := `
		SELECT
			c.name,
			SUM(IF(spender_id = ?, amount - owed, owed)) as user_total,
			SUM(IF(owed > 0, amount, 0)) as couple_total
		FROM outgoings o
		JOIN categories c ON o.category_id=c.id
		WHERE spender_id in (?, ?)
		AND DATE_FORMAT(timestamp,'%Y-%m') = ?
		group by 1
		order by 2 desc;
	`

	rows, err := self.dbh.Query(query, self.ID, self.ID, partnerID, monthYear)

	defer rows.Close()

	if err != nil {
		return nil, err
	}

	var totals []CategoryTotals

	for rows.Next() {
		var c CategoryTotals
		if err := rows.Scan(&c.Category, &c.UserTotal, &c.CoupleTotal); err != nil {
			return nil, err
		}
		totals = append(totals, c)
	}

	return totals, err
}
