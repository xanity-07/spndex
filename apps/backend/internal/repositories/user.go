package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xanity-07/spndex/internal/model"
	"github.com/xanity-07/spndex/internal/model/user"
	"github.com/xanity-07/spndex/internal/server"
)

type UserRepository struct {
	server *server.Server
}

func NewUserRepository(s *server.Server) *UserRepository {
	return &UserRepository{
		server: s,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, payload *user.CreateUserPayload) (*user.User, error) {
	stmt := `
		INSERT INTO
			users (
				id,
				first_name,
				last_name,
				email,
				password_hash,
				created_at,
				updated_at
			)
			VALUES
			(
				@id,
				@first_name,
				@last_name,
				@email,
				@password_hash,
				@created_at,
				@updated_at
			)
		RETURNING
		*
	`
	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id":            uuid.New(),
		"first_name":    payload.FirstName,
		"last_name":     payload.LastName,
		"email":         payload.Email,
		"password_hash": payload.Password,
		"created_at":    time.Now(),
		"updated_at":    time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute create user query: %w", err)
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[user.User])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table users: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) CheckUserExists(ctx context.Context, email string) (*user.User, error) {
	stmt := `
		SELECT
			*
		FROM
			users
		WHERE
			email = @email
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"email": email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute check user exists query email=%s: %w", email, err)
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[user.User])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table users email=%s: %w", email, err)
	}

	return &user, nil
}

func (r *UserRepository) GetUsers(ctx context.Context, query *user.GetUsersQuery) (*model.PaginatedResponse[user.User], error) {
	stmt := `
		SELECT
			id,
			first_name,
			last_name,
			email,
			password_hash,
			created_at,
			updated_at
		FROM
			users
	`
	args := pgx.NamedArgs{}
	conditions := []string{}

	if query.Search != nil {
		conditions = append(conditions, "first_name ILIKE @search OR last_name ILIKE @search OR email ILIKE @search")
		args["search"] = "%" + *query.Search + "%"
	}

	if len(conditions) > 0 {
		stmt += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	countStmt := "SELECT COUNT(*) FROM users"
	err := r.server.DB.Pool.QueryRow(ctx, countStmt, args).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get users count")
	}

	stmt += " OFFSET @offset LIMIT @limit"
	args["limit"] = query.Limit
	args["offset"] = (*query.Page - 1) * *query.Limit

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get users query: %w", err)
	}

	userList, err := pgx.CollectRows(rows, pgx.RowToStructByName[user.User])
	if err != nil {
		// If no rows return a empty paginated response
		if errors.Is(err, pgx.ErrNoRows) {
			return &model.PaginatedResponse[user.User]{
				Data:       []user.User{},
				Page:       *query.Page,
				Limit:      *query.Limit,
				Total:      0,
				TotalPages: 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to collect rows from table users: %w", err)
	}
	return &model.PaginatedResponse[user.User]{
		Data:       userList,
		Page:       *query.Page,
		Limit:      *query.Limit,
		Total:      total,
		TotalPages: (*query.Limit - 1) / *query.Limit,
	}, nil
}
