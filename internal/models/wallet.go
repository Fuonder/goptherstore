package models

import "time"

type MartUserWallet struct {
	ID            int       `json:"-"`
	OwnerID       int       `json:"-"`
	Balance       float32   `json:"current"`
	TotalWithdraw float32   `json:"withdrawn"`
	CreatedAt     time.Time `json:"-"`
}

type Withdrawal struct {
	ID        int       `json:"-"`
	UserID    int       `json:"-"`
	OrderID   string    `json:"order"`
	Amount    float32   `json:"sum"`
	Status    string    `json:"-"`
	CreatedAt time.Time `json:"processed_at,omitempty"`
}
