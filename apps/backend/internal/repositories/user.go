package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xanity-07/spndex/internal/errs"
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
				password_hash
			)
			VALUES
			(
				@id,
				@first_name,
				@last_name,
				@email,
				@password_hash
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

func (r *UserRepository) CheckUserExists(ctx context.Context, email string) (bool, error) {
	stmt := `
	SELECT EXISTS (
    	SELECT 1
   		FROM
			users
    	WHERE
			email = @email
     		AND deleted_at IS NULL
	)
	`
	var exists bool
	err := r.server.DB.Pool.QueryRow(ctx, stmt, pgx.NamedArgs{"email": email}).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *UserRepository) GetUsers(ctx context.Context, query *user.GetUsersQuery) (*model.PaginatedResponse[user.User], error) {
	stmt := `
		SELECT
			id,
			first_name,
			last_name,
			email,
			password_hash,
			user_role,
			created_at,
			updated_at,
			deleted_at
		FROM
			users
	`
	args := pgx.NamedArgs{}
	conditions := []string{"deleted_at IS NULL "}

	if query.Search != nil {
		conditions = append(conditions, "first_name ILIKE @search OR last_name ILIKE @search OR email ILIKE @search")
		args["search"] = "%" + *query.Search + "%"
	}

	if len(conditions) > 0 {
		stmt += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	countStmt := "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL"
	err := r.server.DB.Pool.QueryRow(ctx, countStmt, args).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get users count")
	}

	if query.Order != nil {
		stmt += "ORDER BY " + *query.Order
		if query.Sort != nil && *query.Sort != "desc" {
			stmt += " DESC "
		} else {
			stmt += " ASC "
		}
	} else {
		stmt += " ORDER BY created_at DESC"
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
		TotalPages: (total + *query.Limit - 1) / *query.Limit,
	}, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	stmt := `
		SELECT
			id,
			first_name,
			last_name,
			email,
			password_hash,
			user_role,
			created_at,
			updated_at,
			deleted_at
		FROM
			users
		WHERE
			email = @email
	`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"email": email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[user.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to collect row from table users: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
	stmt := `
		SELECT
			id,
			first_name,
			last_name,
			email,
			password_hash,
			user_role,
			created_at,
			updated_at,
			deleted_at
		FROM
			users
		WHERE
			id = @id
			AND deleted_at IS NULL
	`
	rows, err := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"id": payload.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute get user by id query %w", err)
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[user.User])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table users %w", err)
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, userID uuid.UUID, payload *user.UpdateUserPayload) (*user.User, error) {
	stmt := `
		UPDATE users SET
	`
	args := pgx.NamedArgs{
		"id": userID,
	}

	setClauses := []string{}

	if payload.Email != nil {
		setClauses = append(setClauses, "email = @email")
		args["email"] = *payload.Email
	}

	if payload.Password != nil {
		setClauses = append(setClauses, "password_hash = @password_hash")
		args["password_hash"] = *payload.Password
	}

	if payload.FirstName != nil {
		setClauses = append(setClauses, "first_name = @first_name")
		args["first_name"] = *payload.FirstName
	}

	if payload.LastName != nil {
		setClauses = append(setClauses, "last_name = @last_name")
		args["last_name"] = *payload.LastName
	}

	if len(setClauses) == 0 {
		return nil, errs.NewBadRequestError("no fields to update", false, nil, nil, nil)
	}

	setClauses = append(setClauses, "updated_at = @updated_at")
	args["updated_at"] = time.Now()
	stmt += strings.Join(setClauses, ", ")
	stmt += " WHERE id = @id AND deleted_at IS NULL RETURNING *"

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute update users query: %w", err)
	}

	updatedTodo, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[user.User])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table users: %w", err)
	}

	return &updatedTodo, nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, payload *user.DeleteUserPayload) error {
	stmt := `
		UPDATE users
		SET
			deleted_at = NOW()
		WHERE
			id = @id
			AND deleted_at IS NULL
	`
	result, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{"id": payload.ID})
	if err != nil {
		return fmt.Errorf("failed to execute delete users query: %w", err)
	}

	if result.RowsAffected() == 0 {
		code := "USER_NOT_FOUND"
		return errs.NewBadRequestError("user not found", false, &code, nil, nil)
	}
	return nil
}
