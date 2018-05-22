package outgoing

import "time"

type Outgoing struct {
	ID          *int       `json:"id"`
	Description string     `json:"description"`
	Amount      float64    `json:"amount,string"`
	Owed        float64    `json:"owed,string"`
	Spender     int        `json:"spender,string"`
	Category    string     `json:"category"`
	Settled     *time.Time `json:"settled"`
	Timestamp   *time.Time `json:"timestamp"`
}
