package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/errs"
	"github.com/xanity-07/spndex/internal/model/session"
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/tests"
)

func TestSessionRepository_Create(t *testing.T) {
	testRedis := tests.NewRedisTestDatabase(t)

	srv := &server.Server{
		Redis: testRedis.Client,
	}

	repo := NewSessionRepository(srv)

	ctx := context.Background()

	sessionID := uuid.NewString()
	userID := uuid.NewString()

	s := &session.Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}

	err := repo.Create(ctx, s, time.Hour)
	if err != nil {
		t.Fatalf("Create returned an error: %v", err)
	}

	key := enums.SessionKeyPrefix.Key(sessionID)

	result, err := testRedis.Client.Get(ctx, key).Bytes()
	if err != nil {
		t.Fatalf("failed to get session directly from redis: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("expected session data to be stored in redis")
	}

	storedTTL, err := testRedis.Client.TTL(ctx, key).Result()
	if err != nil {
		t.Fatalf("failed to get session TTL: %v", err)
	}

	if storedTTL <= 0 {
		t.Fatal("expected session to have a positive TTL")
	}

	if storedTTL > time.Hour {
		t.Errorf("expected TTL to be at most 1 hour, got %v", storedTTL)
	}
}

func TestSessionRepository_Get(t *testing.T) {
	testRedis := tests.NewRedisTestDatabase(t)

	srv := &server.Server{
		Redis: testRedis.Client,
	}

	repo := NewSessionRepository(srv)

	ctx := context.Background()

	sessionID := uuid.NewString()
	userID := uuid.NewString()

	expected := &session.Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}

	err := repo.Create(ctx, expected, time.Hour)
	if err != nil {
		t.Fatalf("Create returned an error: %v", err)
	}

	result, err := repo.Get(ctx, sessionID)
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}

	if result == nil {
		t.Fatal("expected session, got nil")
	}

	if result.ID != expected.ID {
		t.Errorf("expected session ID %q, got %q", expected.ID, result.ID)
	}

	if result.UserID != expected.UserID {
		t.Errorf("expected user ID %q, got %q", expected.UserID, result.UserID)
	}

	if !result.CreatedAt.Equal(expected.CreatedAt) {
		t.Errorf(
			"expected created at %v, got %v",
			expected.CreatedAt,
			result.CreatedAt,
		)
	}
}

func TestSessionRepository_Get_NotFound(t *testing.T) {
	testRedis := tests.NewRedisTestDatabase(t)

	srv := &server.Server{
		Redis: testRedis.Client,
	}

	repo := NewSessionRepository(srv)

	ctx := context.Background()

	result, err := repo.Get(ctx, uuid.NewString())

	if result != nil {
		t.Errorf("expected nil session, got %+v", result)
	}

	if !errors.Is(err, errs.ErrSessionNotFound) {
		t.Fatalf(
			"expected ErrSessionNotFound, got %v",
			err,
		)
	}
}

func TestSessionRepository_Create_TTL(t *testing.T) {
	testRedis := tests.NewRedisTestDatabase(t)

	srv := &server.Server{
		Redis: testRedis.Client,
	}

	repo := NewSessionRepository(srv)

	ctx := context.Background()

	sessionID := uuid.NewString()

	s := &session.Session{
		ID:        sessionID,
		UserID:    uuid.NewString(),
		CreatedAt: time.Now().UTC(),
	}

	ttl := 2 * time.Second

	err := repo.Create(ctx, s, ttl)
	if err != nil {
		t.Fatalf("Create returned an error: %v", err)
	}

	key := enums.SessionKeyPrefix.Key(sessionID)

	storedTTL, err := testRedis.Client.TTL(ctx, key).Result()
	if err != nil {
		t.Fatalf("failed to get session TTL: %v", err)
	}

	if storedTTL <= 0 {
		t.Fatal("expected session to have a positive TTL")
	}

	if storedTTL > ttl {
		t.Errorf(
			"expected TTL to be at most %v, got %v",
			ttl,
			storedTTL,
		)
	}

	time.Sleep(ttl + 500*time.Millisecond)

	result, err := repo.Get(ctx, sessionID)

	if result != nil {
		t.Errorf("expected nil session after expiration, got %+v", result)
	}

	if !errors.Is(err, errs.ErrSessionNotFound) {
		t.Fatalf(
			"expected ErrSessionNotFound after expiration, got %v",
			err,
		)
	}
}

func TestSessionRepository_Delete(t *testing.T) {
	testRedis := tests.NewRedisTestDatabase(t)

	srv := &server.Server{
		Redis: testRedis.Client,
	}

	repo := NewSessionRepository(srv)

	ctx := context.Background()

	sessionID := uuid.NewString()

	s := &session.Session{
		ID:        sessionID,
		UserID:    uuid.NewString(),
		CreatedAt: time.Now().UTC(),
	}

	err := repo.Create(ctx, s, time.Hour)
	if err != nil {
		t.Fatalf("Create returned an error: %v", err)
	}

	err = repo.Delete(ctx, sessionID)
	if err != nil {
		t.Fatalf("Delete returned an error: %v", err)
	}

	result, err := repo.Get(ctx, sessionID)

	if result != nil {
		t.Errorf("expected nil session after deletion, got %+v", result)
	}

	if !errors.Is(err, errs.ErrSessionNotFound) {
		t.Fatalf(
			"expected ErrSessionNotFound after deletion, got %v",
			err,
		)
	}
}
