package repository

import (
	"context"
	"database/sql"
	"errors"
	"go-crud/internal/entity"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryRepository interface {
	CreateRepository(ctx context.Context, repo *entity.Repository) error
	GetRepositoryByID(ctx context.Context, id int) (*entity.Repository, error)
	GetByID(ctx context.Context, id int) (*entity.Repository, error)   // ✅ Tambahkan GetByID
	Update(ctx context.Context, repo *entity.Repository) error         // ✅ Tambahkan Update
	Delete(ctx context.Context, id int) error                          // ✅ Tambahkan Delete
}


type repoRepository struct {
	db  *pgxpool.Pool 
}

func NewRepositoryRepository(db  *pgxpool.Pool ) RepositoryRepository {
	return &repoRepository{db: db}
}

func (r *repoRepository) CreateRepository(ctx context.Context, repo *entity.Repository) error {
	query := "INSERT INTO repositories (user_id, name, url, ai_enabled, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING id"
	err := r.db.QueryRow(ctx, query, repo.UserID, repo.Name, repo.URL, repo.AIEnabled).Scan(&repo.ID)
	return err
}

func (r *repoRepository) GetRepositoryByID(ctx context.Context, id int) (*entity.Repository, error) {
	query := "SELECT id, user_id, name, url, ai_enabled, created_at, updated_at FROM repositories WHERE id = $1"
	row := r.db.QueryRow(ctx, query, id)

	var repo entity.Repository
	err := row.Scan(&repo.ID, &repo.UserID, &repo.Name, &repo.URL, &repo.AIEnabled, &repo.CreatedAt, &repo.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("repository not found")
		}
		return nil, err
	}
	return &repo, nil
}

func (r *repoRepository) GetByID(ctx context.Context, id int) (*entity.Repository, error) {
	query := "SELECT id, user_id, name, url, ai_enabled, created_at, updated_at FROM repositories WHERE id = $1"
	row := r.db.QueryRow(ctx, query, id)

	var repo entity.Repository
	err := row.Scan(&repo.ID, &repo.UserID, &repo.Name, &repo.URL, &repo.AIEnabled, &repo.CreatedAt, &repo.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("repository not found")
		}
		return nil, err
	}
	return &repo, nil
}

func (r *repoRepository) Update(ctx context.Context, repo *entity.Repository) error {
	query := "UPDATE repositories SET name = $1, url = $2, ai_enabled = $3, updated_at = NOW() WHERE id = $4"
	_, err := r.db.Exec(ctx, query, repo.Name, repo.URL, repo.AIEnabled, repo.ID)
	return err
}

func (r *repoRepository) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM repositories WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	return err
}
