package models

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type MartUser struct {
	ID        int       `json:"id"`
	Login     string    `json:"login"`
	Password  string    `json:"pwd"`
	CreatedAt time.Time `json:"created_at"`
}
