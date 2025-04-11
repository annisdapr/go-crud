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

type RepositoryRepository interface {
	CreateRepository(ctx context.Context, repo *entity.Repository) error
	GetRepositoryByID(ctx context.Context, id int) (*entity.Repository, error)
	GetAllRepositories(ctx context.Context) ([]entity.Repository, error)

	GetByID(ctx context.Context, id int) (*entity.Repository, error)
	GetRepositoriesByUserID(ctx context.Context, userID int) ([]entity.Repository, error) 
	Update(ctx context.Context, repo *entity.Repository) error
	Delete(ctx context.Context, id int) error
}

type repoRepository struct {
	db  *pgxpool.Pool 
}

func NewRepositoryRepository(db  *pgxpool.Pool ) RepositoryRepository {
	return &repoRepository{db: db}
}

func (r *repoRepository) CreateRepository(ctx context.Context, repo *entity.Repository) error {
	ctx, span := tracing.Tracer.Start(ctx, "repoRepository.CreateRepository")
	defer span.End()

	query := `INSERT INTO repositories (user_id, name, url, ai_enabled, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING id`

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "INSERT"),
		attribute.String("db.statement", query),
		attribute.Int("db.user_id", repo.UserID),
		attribute.String("db.repo_name", repo.Name),
	)

	err := r.db.QueryRow(ctx, query, repo.UserID, repo.Name, repo.URL, repo.AIEnabled).Scan(&repo.ID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Log ID yang baru dibuat
	span.SetAttributes(attribute.Int("db.generated_id", repo.ID))

	return nil
}

func (r *repoRepository) GetRepositoriesByUserID(ctx context.Context, userID int) ([]entity.Repository, error) {
	ctx, span := tracing.Tracer.Start(ctx, "repoRepository.GetRepositoriesByUserID")
	defer span.End()

	query := "SELECT id, user_id, name, url, ai_enabled, created_at, updated_at FROM repositories WHERE user_id = $1"

	// Set attributes awal
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.statement", query),
		attribute.Int("db.user_id", userID),
	)

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var repositories []entity.Repository
	for rows.Next() {
		var repo entity.Repository
		err := rows.Scan(&repo.ID, &repo.UserID, &repo.Name, &repo.URL, &repo.AIEnabled, &repo.CreatedAt, &repo.UpdatedAt)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		repositories = append(repositories, repo)
	}

	if len(repositories) == 0 {
		err := errors.New("no repositories found for this user")
		span.RecordError(err)
		return nil, err
	}
	span.SetAttributes(attribute.Int("db.response_count", len(repositories)))

	return repositories, nil
}


func (r *repoRepository) GetRepositoryByID(ctx context.Context, id int) (*entity.Repository, error) {
	ctx, span := tracing.Tracer.Start(ctx, "repoRepository.GetRepositoryByID")
	defer span.End()

	query := "SELECT id, user_id, name, url, ai_enabled, created_at, updated_at FROM repositories WHERE id = $1"
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.statement", query),
		attribute.Int("db.repository_id", id),
	)

	row := r.db.QueryRow(ctx, query, id)

	var repo entity.Repository
	err := row.Scan(&repo.ID, &repo.UserID, &repo.Name, &repo.URL, &repo.AIEnabled, &repo.CreatedAt, &repo.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			span.RecordError(err)
			return nil, errors.New("repository not found")
		}
		span.RecordError(err)
		return nil, err
	}
	//hasil query
	span.SetAttributes(
		attribute.Int("db.result.repo_id", repo.ID),
		attribute.String("db.result.repo_name", repo.Name),
		attribute.String("db.result.repo_url", repo.URL),
	)

	return &repo, nil
}


func (r *repoRepository) GetAllRepositories(ctx context.Context) ([]entity.Repository, error) {
	ctx, span := tracing.Tracer.Start(ctx, "repoRepository.GetAllRepositories")
	defer span.End()

	query := "SELECT id, user_id, name, url, ai_enabled, created_at, updated_at FROM repositories"
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

	var repositories []entity.Repository
	for rows.Next() {
		var repo entity.Repository
		err := rows.Scan(&repo.ID, &repo.UserID, &repo.Name, &repo.URL, &repo.AIEnabled, &repo.CreatedAt, &repo.UpdatedAt)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		repositories = append(repositories, repo)
	}

	span.SetAttributes(
		attribute.Int("db.result.count", len(repositories)),
	)

	return repositories, nil
}


func (r *repoRepository) GetByID(ctx context.Context, id int) (*entity.Repository, error) {
	ctx, span := tracing.Tracer.Start(ctx, "repoRepository.GetByID")
	defer span.End()

	query := "SELECT id, user_id, name, url, ai_enabled, created_at, updated_at FROM repositories WHERE id = $1"
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.statement", query),
		attribute.Int("db.repository_id", id),
	)

	row := r.db.QueryRow(ctx, query, id)

	var repo entity.Repository
	err := row.Scan(&repo.ID, &repo.UserID, &repo.Name, &repo.URL, &repo.AIEnabled, &repo.CreatedAt, &repo.UpdatedAt)
	if err != nil {
		span.RecordError(err)
		if errors.Is(err, sql.ErrNoRows) {
			span.SetAttributes(attribute.String("db.result", "not found"))
			return nil, errors.New("repository not found")
		}
		return nil, err
	}

	span.SetAttributes(
		attribute.String("db.result.repository_name", repo.Name),
		attribute.Int("db.result.repository_user_id", repo.UserID),
	)

	return &repo, nil
}


func (r *repoRepository) Update(ctx context.Context, repo *entity.Repository) error {
	ctx, span := tracing.Tracer.Start(ctx, "RepoRepository.Update")
	defer span.End()

	// Ambil data sebelum update untuk keperluan audit perubahan
	var oldRepo entity.Repository
	querySelect := "SELECT name, url, ai_enabled FROM repositories WHERE id = $1"
	err := r.db.QueryRow(ctx, querySelect, repo.ID).Scan(&oldRepo.Name, &oldRepo.URL, &oldRepo.AIEnabled)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Tracing query update
	queryUpdate := "UPDATE repositories SET name = $1, url = $2, ai_enabled = $3, updated_at = NOW() WHERE id = $4"
	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "UPDATE"),
		attribute.String("db.statement", queryUpdate),
		attribute.Int("db.repository_id", repo.ID),
		attribute.String("db.old.name", oldRepo.Name),
		attribute.String("db.old.url", oldRepo.URL),
		attribute.Bool("db.old.ai_enabled", oldRepo.AIEnabled),
		attribute.String("db.new.name", repo.Name),
		attribute.String("db.new.url", repo.URL),
		attribute.Bool("db.new.ai_enabled", repo.AIEnabled),
	)

	_, err = r.db.Exec(ctx, queryUpdate, repo.Name, repo.URL, repo.AIEnabled, repo.ID)
	if err != nil {
		span.RecordError(err)
	}
	return err
}


func (r *repoRepository) Delete(ctx context.Context, id int) error {
	ctx, span := tracing.Tracer.Start(ctx, "repoRepository.Delete")
	defer span.End()

	query := "DELETE FROM repositories WHERE id = $1"

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "DELETE"),
		attribute.String("db.statement", query),
		attribute.Int("db.repository_id", id),
	)

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		span.RecordError(err)
	}
	return err
}

