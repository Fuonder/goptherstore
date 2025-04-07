package accrualservice

import "errors"

var (
	ErrNotRegistered            = errors.New("order is not registered")
	ErrToManyRequests           = errors.New("too many requests")
	ErrInternalServerError      = errors.New("internal server error")
	ErrCanNotGetAccrualResponse = errors.New("can not get accrual response")
)

var (
	AccrualStatusRegistered = "REGISTERED"
	AccrualStatusInvalid    = "INVALID"
	AccrualStatusProcessing = "PROCESSING"
	AccrualStatusProcessed  = "PROCESSED"
)

type AccrualService interface {
	Run() error
}
