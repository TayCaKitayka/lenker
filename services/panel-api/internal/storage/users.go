package storage

import (
	"context"
	"database/sql"
)

type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Status      string `json:"status"`
	DisplayName string `json:"display_name"`
}

type UsersRepository interface {
	List(ctx context.Context) ([]User, error)
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
