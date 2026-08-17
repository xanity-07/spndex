package authmodel

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}
