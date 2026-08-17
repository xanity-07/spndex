package session

import (
	"time"
)

type Session struct {
	CreatedAt time.Time `json:"createdAt"`
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
}
