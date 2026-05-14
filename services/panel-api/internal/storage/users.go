package storage

import (
	"context"
	"database/sql"
	"errors"
)

type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Status      string `json:"status"`
	DisplayName string `json:"display_name"`
}

type CreateUserInput struct {
	Email       string
	DisplayName string
}

type UpdateUserInput struct {
	Email       *string
	DisplayName *string
}

type UsersRepository interface {
	List(ctx context.Context) ([]User, error)
	Create(ctx context.Context, input CreateUserInput) (User, error)
	FindByID(ctx context.Context, id string) (User, error)
	Update(ctx context.Context, id string, input UpdateUserInput) (User, error)
	SetStatus(ctx context.Context, id string, status string) (User, error)
}

type usersRepository struct {
	db *sql.DB
}

func NewUsersRepository(db *sql.DB) UsersRepository {
	return &usersRepository{db: db}
}

func (r *usersRepository) List(ctx context.Context) ([]User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, email, status, display_name
		FROM users
		ORDER BY created_at DESC
		LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Email, &user.Status, &user.DisplayName); err != nil {
			return nil, err
		}
		result = append(result, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *usersRepository) Create(ctx context.Context, input CreateUserInput) (User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (email, display_name)
		VALUES ($1, $2)
		RETURNING id::text, email, status, display_name
	`, input.Email, input.DisplayName).Scan(&user.ID, &user.Email, &user.Status, &user.DisplayName)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *usersRepository) FindByID(ctx context.Context, id string) (User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, `
		SELECT id::text, email, status, display_name
		FROM users
		WHERE id = $1
	`, id).Scan(&user.ID, &user.Email, &user.Status, &user.DisplayName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return user, nil
}

func (r *usersRepository) Update(ctx context.Context, id string, input UpdateUserInput) (User, error) {
	current, err := r.FindByID(ctx, id)
	if err != nil {
		return User{}, err
	}
	if input.Email != nil {
		current.Email = *input.Email
	}
	if input.DisplayName != nil {
		current.DisplayName = *input.DisplayName
	}

	var user User
	err = r.db.QueryRowContext(ctx, `
		UPDATE users
		SET email = $2, display_name = $3, updated_at = now()
		WHERE id = $1
		RETURNING id::text, email, status, display_name
	`, id, current.Email, current.DisplayName).Scan(&user.ID, &user.Email, &user.Status, &user.DisplayName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return user, nil
}

func (r *usersRepository) SetStatus(ctx context.Context, id string, status string) (User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, `
		UPDATE users
		SET status = $2, updated_at = now()
		WHERE id = $1
		RETURNING id::text, email, status, display_name
	`, id, status).Scan(&user.ID, &user.Email, &user.Status, &user.DisplayName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return user, nil
}
