package models

import "errors"

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserCreationFailed = errors.New("user creation failed")
	ErrWrongCredentials   = errors.New("wrong credentials")

	ErrOrderAlreadyExists = errors.New("order already exists")
	ErrOrderOfOtherUser   = errors.New("order already registered by other user")
	ErrInvalidOrderNumber = errors.New("invalid order number")
	
	ErrNotEnoughBonuses = errors.New("not enough bonuses")

	ErrNoData = errors.New("no data")
)
