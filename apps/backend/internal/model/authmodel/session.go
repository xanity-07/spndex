package authmodel

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	CreatedAt time.Time
	ExpiresAt time.Time
	ID        uuid.UUID
	UserID    uuid.UUID
}
