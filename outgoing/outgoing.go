package outgoing

import "time"

type Outgoing struct {
	ID          int        `json:"id"`
	Description string     `json:"description"`
	Amount      float64    `json:"amount"`
	Spender     int        `json:"spender"`
	Category    string     `json:"category"`
	Settled     *time.Time `json:"settled"`
	Timestamp   time.Time  `json:"timestamp"`
}
