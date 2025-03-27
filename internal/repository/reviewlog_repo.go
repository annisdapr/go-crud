package repository

import (
	"context"
	"errors"
	"go-crud/internal/entity"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CodeReviewRepository interface {
	InsertCodeReviewLog(ctx context.Context, log *entity.CodeReviewLog) error
	GetCodeReviewLogsByRepoID(ctx context.Context, repoID int) ([]entity.CodeReviewLog, error)
}

type codeReviewRepository struct {
	db *pgxpool.Pool
}

func NewCodeReviewRepository(db *pgxpool.Pool) CodeReviewRepository {
	return &codeReviewRepository{db: db}
}

// Insert log review
func (r *codeReviewRepository) InsertCodeReviewLog(ctx context.Context, log *entity.CodeReviewLog) error {
	// Gunakan context dengan timeout agar tidak menggantung
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `INSERT INTO codereview_log (repository_id, review_result, created_at) 
              VALUES ($1, $2, NOW()) RETURNING id`

	err := r.db.QueryRow(ctx, query, log.RepositoryID, log.ReviewResult).Scan(&log.ID)
	if err != nil {
		return err
	}

	return nil
}

// Get logs by repository ID
func (r *codeReviewRepository) GetCodeReviewLogsByRepoID(ctx context.Context, repoID int) ([]entity.CodeReviewLog, error) {
	// Gunakan context dengan timeout agar tidak menggantung
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := "SELECT id, repository_id, review_result, created_at FROM codereview_log WHERE repository_id = $1"
	rows, err := r.db.Query(ctx, query, repoID)
	if err != nil {
		// Jika tidak ada data, kembalikan error yang lebih deskriptif
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("🚫 no review logs found for the given repository")
		}
		return nil, err
	}
	defer rows.Close()

	var logs []entity.CodeReviewLog
	for rows.Next() {
		var log entity.CodeReviewLog
		if err := rows.Scan(&log.ID, &log.RepositoryID, &log.ReviewResult, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	// Periksa jika terjadi error selama iterasi rows
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Jika logs kosong, kembalikan error yang sesuai
	if len(logs) == 0 {
		return nil, errors.New("🚫 no review logs found for the given repository")
	}

	return logs, nil
}
