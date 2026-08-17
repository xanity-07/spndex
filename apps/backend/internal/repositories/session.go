package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/errs"
	"github.com/xanity-07/spndex/internal/model/session"
	"github.com/xanity-07/spndex/internal/server"
)

type SessionInterface interface {
	Create(ctx context.Context, session *session.Session, ttl time.Duration) error
	Get(ctx context.Context, sessionID string) (*session.Session, error)
	Delete(ctx context.Context, sessionID string) error
}

type SessionRepository struct {
	server *server.Server
}

func NewSessionRepository(s *server.Server) *SessionRepository {
	return &SessionRepository{
		server: s,
	}
}

func (r *SessionRepository) Create(ctx context.Context, session *session.Session, ttl time.Duration) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := enums.SessionKeyPrefix.Key(session.ID)
	if err := r.server.Redis.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to store session in redis: %T", err)
	}

	return nil
}

func (r *SessionRepository) Get(ctx context.Context, sessionID string) (*session.Session, error) {
	data, err := r.server.Redis.Get(ctx, enums.SessionKeyPrefix.Key(sessionID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errs.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session from redis: %w", err)
	}

	var session session.Session
	err = json.Unmarshal(data, &session)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

func (r *SessionRepository) Delete(ctx context.Context, sessionID string) error {
	if err := r.server.Redis.Del(ctx, enums.SessionKeyPrefix.Key(sessionID)).Err(); err != nil {
		return fmt.Errorf("failed to delete session from redis: %w", err)
	}
	return nil
}
