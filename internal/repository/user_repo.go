package repository

import (
	"context"
	"database/sql"
	"errors"
	"go-crud/internal/entity"
	"go-crud/internal/tracing"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
)

// UserRepository interface
type UserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByID(ctx context.Context, id int) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error 
	DeleteUser(ctx context.Context, id int) error    
	GetAllUsers(ctx context.Context) ([]entity.User, error)     
	GetByEmail(ctx context.Context, email string) (*entity.User, error) 
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
	ctx, span := tracing.Tracer.Start(ctx, "userRepository.CreateUser")
	defer span.End()

	query := "INSERT INTO users (name, email, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id"

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "INSERT"),
		attribute.String("db.statement", query),
		attribute.String("db.user.name", user.Name),
		attribute.String("db.user.email", user.Email),
	)

	err := r.db.QueryRow(ctx, query, user.Name, user.Email).Scan(&user.ID)
	if err != nil {
		span.RecordError(err)
	} else {
		span.SetAttributes(attribute.Int("db.user.id", user.ID))
	}

	return err
}


func (r *userRepository) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	ctx, span := tracing.Tracer.Start(ctx, "userRepository.GetUserByID")
	defer span.End()

	query := "SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1"

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.statement", query),
		attribute.Int("db.user.id.param", id),
	)

	row := r.db.QueryRow(ctx, query, id)

	var user entity.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		span.RecordError(err)
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Catat hasil jika berhasil
	span.SetAttributes(
		attribute.Int("db.user.id.result", user.ID),
		attribute.String("db.user.name", user.Name),
		attribute.String("db.user.email", user.Email),
	)

	return &user, nil
}

// UpdateUser memperbarui data user
func (r *userRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	ctx, span := tracing.Tracer.Start(ctx, "userRepository.UpdateUser")
	defer span.End()

	// Ambil data lama sebelum update
	oldUser, err := r.GetUserByID(ctx, user.ID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Buat query update
	query := "UPDATE users SET name = $1, email = $2, updated_at = NOW() WHERE id = $3"

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "UPDATE"),
		attribute.String("db.statement", query),
		attribute.Int("db.user.id", user.ID),
		attribute.String("db.user.name.old", oldUser.Name),
		attribute.String("db.user.name.new", user.Name),
		attribute.String("db.user.email.old", oldUser.Email),
		attribute.String("db.user.email.new", user.Email),
	)

	_, err = r.db.Exec(ctx, query, user.Name, user.Email, user.ID)
	if err != nil {
		span.RecordError(err)
	}
	return err
}


// DeleteUser menghapus user berdasarkan ID
func (r *userRepository) DeleteUser(ctx context.Context, id int) error {
	ctx, span := tracing.Tracer.Start(ctx, "userRepository.DeleteUser")
	defer span.End()

	query := "DELETE FROM users WHERE id = $1"

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "DELETE"),
		attribute.String("db.statement", query),
		attribute.Int("db.user.id", id),
	)

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		span.RecordError(err)
	}
	return err
}

// GetAllUsers mengambil semua user dari database
func (r *userRepository) GetAllUsers(ctx context.Context) ([]entity.User, error) {
	ctx, span := tracing.Tracer.Start(ctx, "UserRepository.GetAllUsers")
	defer span.End()

	query := "SELECT id, name, email, created_at, updated_at FROM users"
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.statement", query),
	)

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			span.RecordError(err)
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(
		attribute.Int("db.result.count", len(users)),
	)

	return users, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	ctx, span := tracing.Tracer.Start(ctx, "userRepository.GetByEmail")
	defer span.End()

	query := `SELECT id, name, email, created_at, updated_at FROM users WHERE email = $1`
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.statement", query),
		attribute.String("db.user.email", email),
	)

	row := r.db.QueryRow(ctx, query, email)

	var user entity.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		span.RecordError(err)
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	span.SetAttributes(attribute.Int("db.user.id.result", user.ID))
	return &user, nil
}



