package repository

import (
	"context"
	"database/sql"
	"errors"
	"go-crud/internal/entity"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository interface
type UserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByID(ctx context.Context, id int) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error // Tambahkan UpdateUser
	DeleteUser(ctx context.Context, id int) error           // Tambahkan DeleteUser
}

type userRepository struct {
	db *pgxpool.Pool 
}

// NewUserRepository membuat instance UserRepository
func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

// CreateUser untuk menyimpan user ke database
func (r *userRepository) CreateUser(ctx context.Context, user *entity.User) error {
	query := "INSERT INTO users (name, email, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id"
	return r.db.QueryRow(ctx, query, user.Name, user.Email).Scan(&user.ID)
}

// GetUserByID mengambil user berdasarkan ID
func (r *userRepository) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	query := "SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1"
	row := r.db.QueryRow(ctx, query, id)

	var user entity.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser memperbarui data user
func (r *userRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	query := "UPDATE users SET name = $1, email = $2, updated_at = NOW() WHERE id = $3"
	_, err := r.db.Exec(ctx, query, user.Name, user.Email, user.ID)
	return err
}

// DeleteUser menghapus user berdasarkan ID
func (r *userRepository) DeleteUser(ctx context.Context, id int) error {
	query := "DELETE FROM users WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	return err
}
