package models

import "time"

var (
	OrderStatusNew        = "NEW"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusProcessed  = "PROCESSED"
)

type MartOrder struct {
	ID        int       `json:"-"`
	UserID    int       `json:"-"`
	OrderID   string    `json:"number"`
	Status    string    `json:"status"`
	Bonus     float32   `json:"accrual,omitempty"`
	CreatedAt time.Time `json:"uploaded_at,omitempty"`
}
