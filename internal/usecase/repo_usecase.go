package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-crud/internal/entity"
	"go-crud/internal/repository"
	"go-crud/internal/tracing"
	"go-crud/internal/circuitbreaker"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
)

// Interface untuk RepositoryUsecase
type IRepositoryUsecase interface {
	CreateRepository(ctx context.Context, repo *entity.Repository) error
	GetRepositoryByID(ctx context.Context, id int) (*entity.Repository, error)
	GetAllRepositories(ctx context.Context) ([]entity.Repository, error)
	UpdateRepository(ctx context.Context, id int, input RepositoryInput) (entity.Repository, error)
	DeleteRepository(ctx context.Context, id int) error
	GetRepositoriesByUserID(ctx context.Context, userID int) ([]entity.Repository, error)
}

// Struct RepositoryUsecase
type RepositoryUsecase struct {
	repoRepo   repository.RepositoryRepository
	userRepo   repository.UserRepository
	redis      *redis.Client
	cbRedis    *gobreaker.CircuitBreaker
	cbPostgres *gobreaker.CircuitBreaker
}

// Input struct untuk repository
type RepositoryInput struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	AIEnabled bool   `json:"ai_enabled"`
}

// NewRepositoryUsecase membuat instance baru
func NewRepositoryUsecase(
	repoRepo repository.RepositoryRepository,
	userRepo repository.UserRepository,
	redisClient *redis.Client,
) IRepositoryUsecase {
	return &RepositoryUsecase{
		repoRepo:   repoRepo,
		userRepo:   userRepo,
		redis:      redisClient,
		cbRedis:    cbreaker.NewBreaker("RepositoryRedisCB"),
		cbPostgres: cbreaker.NewBreaker("RepositoryPostgresCB"),
	}
}

// ✅ Create hanya validasi user dan kembalikan nil untuk Kafka layer
func (u *RepositoryUsecase) CreateRepository(ctx context.Context, repo *entity.Repository) error {
	ctx, span := tracing.Tracer.Start(ctx, "RepositoryUsecase.CreateRepository")
	defer span.End()

	_, err := u.userRepo.GetUserByID(ctx, repo.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	// Tidak langsung simpan ke DB, tugas Kafka consumer
	return nil
}

// ✅ Ambil semua repo
func (u *RepositoryUsecase) GetAllRepositories(ctx context.Context) ([]entity.Repository, error) {
	ctx, span := tracing.Tracer.Start(ctx, "RepositoryUsecase.GetAllRepositories")
	defer span.End()

	return u.repoRepo.GetAllRepositories(ctx)
}

// ✅ Ambil semua repo berdasarkan user
func (u *RepositoryUsecase) GetRepositoriesByUserID(ctx context.Context, userID int) ([]entity.Repository, error) {
	ctx, span := tracing.Tracer.Start(ctx, "RepositoryUsecase.GetRepositoriesByUserID")
	defer span.End()

	_, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return nil, errors.New("user not found")
	}

	return u.repoRepo.GetRepositoriesByUserID(ctx, userID)
}

// ✅ Get by ID, coba cache, fallback ke DB
func (u *RepositoryUsecase) GetRepositoryByID(ctx context.Context, id int) (*entity.Repository, error) {
	ctx, span := tracing.Tracer.Start(ctx, "RepositoryUsecase.GetRepositoryByID")
	defer span.End()

	cacheKey := fmt.Sprintf("repository:%d", id)

	val, err := u.cbRedis.Execute(func() (interface{}, error) {
		return u.redis.Get(ctx, cacheKey).Result()
	})
	if err == nil {
		var cached entity.Repository
		if unmarshalErr := json.Unmarshal([]byte(val.(string)), &cached); unmarshalErr == nil {
			return &cached, nil
		}
	}

	// Ambil dari DB
	result, err := u.cbPostgres.Execute(func() (interface{}, error) {
		return u.repoRepo.GetRepositoryByID(ctx, id)
	})
	if err != nil {
		return nil, err
	}

	return result.(*entity.Repository), nil
}

// ✅ Update hanya ubah data, tidak push Kafka/cache
func (u *RepositoryUsecase) UpdateRepository(ctx context.Context, id int, input RepositoryInput) (entity.Repository, error) {
	ctx, span := tracing.Tracer.Start(ctx, "RepositoryUsecase.UpdateRepository")
	defer span.End()

	repo, err := u.repoRepo.GetRepositoryByID(ctx, id)
	if err != nil {
		return entity.Repository{}, err
	}

	repo.Name = input.Name
	repo.URL = input.URL
	repo.AIEnabled = input.AIEnabled
	repo.UpdatedAt = time.Now()

	if err := u.repoRepo.Update(ctx, repo); err != nil {
		return entity.Repository{}, err
	}

	return *repo, nil
}

// ✅ Delete hanya validasi, tanpa DB call langsung
func (u *RepositoryUsecase) DeleteRepository(ctx context.Context, id int) error {
	_, span := tracing.Tracer.Start(ctx, "RepositoryUsecase.DeleteRepository")
	defer span.End()

	// Tidak langsung hapus dari DB (tugas consumer)
	return nil
}
