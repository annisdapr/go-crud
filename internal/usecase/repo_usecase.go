package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-crud/internal/entity"
	"go-crud/internal/kafka"
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
	repoRepo repository.RepositoryRepository
	userRepo repository.UserRepository
	kafkaProducer *kafka.KafkaProducer
	redisClient   *redis.Client
	cbRedis       *gobreaker.CircuitBreaker
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
	kafkaProducer *kafka.KafkaProducer,
	redisClient *redis.Client,
) IRepositoryUsecase {
	return &RepositoryUsecase{
		repoRepo:      repoRepo,
		userRepo:      userRepo,
		kafkaProducer: kafkaProducer,
		redisClient:   redisClient,
		cbRedis: cbreaker.NewBreaker("RepositoryRedisCB"),
	}
}


func (u *RepositoryUsecase) CreateRepository(ctx context.Context, repo *entity.Repository) error {
	ctx, span := tracing.Tracer.Start(ctx, "CreateRepository")
	defer span.End()

	_, err := u.userRepo.GetUserByID(ctx, repo.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	err = u.repoRepo.CreateRepository(ctx, repo)
	if err != nil {
		return err
	}

	event := map[string]interface{}{
		"event":    "repository.created",
		"id":       repo.ID,
		"user_id":  repo.UserID,
		"name":     repo.Name,
		"url":      repo.URL,
		"ai_enabled": repo.AIEnabled,
	}
	go u.kafkaProducer.Publish(event, "repository.created")

	return nil
}


func (u *RepositoryUsecase) GetAllRepositories(ctx context.Context) ([]entity.Repository, error) {
	ctx, span := tracing.Tracer.Start(ctx, "GetAllRepositories") 
	defer span.End()
	return u.repoRepo.GetAllRepositories(ctx)
}

func (u *RepositoryUsecase) GetRepositoriesByUserID(ctx context.Context, userID int) ([]entity.Repository, error) {
	ctx, span := tracing.Tracer.Start(ctx, "GetRepositoriesByUserID") 
	defer span.End()
	
	// Pastikan user ada sebelum mengambil repositorinya
	_, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		span.RecordError(err) 
		return nil, errors.New("user not found")
	}

	repos, err := u.repoRepo.GetRepositoriesByUserID(ctx, userID)
	if err != nil {
		span.RecordError(err) 
		return nil, err
	}

	return repos, nil
}


func (u *RepositoryUsecase) GetRepositoryByID(ctx context.Context, id int) (*entity.Repository, error) {
	ctx, span := tracing.Tracer.Start(ctx, "GetRepositoryByID") 
	defer span.End()

	cacheKey := fmt.Sprintf("repository:%d", id)

	val, err := u.cbRedis.Execute(func() (interface{}, error) {
		return u.redisClient.Get(ctx, cacheKey).Result()
	})

	if err == nil {
		var cachedRepo entity.Repository
		if unmarshalErr := json.Unmarshal([]byte(val.(string)), &cachedRepo); unmarshalErr == nil {
			return &cachedRepo, nil
		}
	}

	// Fallback ke database
	repo, err := u.repoRepo.GetRepositoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	repoJSON, _ := json.Marshal(repo)
	_, _ = u.cbRedis.Execute(func() (interface{}, error) {
		return nil, u.redisClient.Set(ctx, cacheKey, repoJSON, 2*time.Minute).Err()
	})

	return repo, nil
}

func (u *RepositoryUsecase) UpdateRepository(ctx context.Context, id int, input RepositoryInput) (entity.Repository, error) {
	ctx, span := tracing.Tracer.Start(ctx, "UpdateRepository")
	defer span.End()

	repo, err := u.repoRepo.GetRepositoryByID(ctx, id)
	if err != nil {
		return entity.Repository{}, err
	}

	repo.Name = input.Name
	repo.URL = input.URL
	repo.AIEnabled = input.AIEnabled

	if err := u.repoRepo.Update(ctx, repo); err != nil {
		return entity.Repository{}, err
	}

	event := map[string]interface{}{
		"event":      "repository.updated",
		"id":         repo.ID,
		"user_id":    repo.UserID,
		"name":       repo.Name,
		"url":        repo.URL,
		"ai_enabled": repo.AIEnabled,
	}
	go u.kafkaProducer.Publish(event, "repository.updated")

	// Hapus cache setelah update dan publish event
	go func() {
		cacheKey := fmt.Sprintf("repository:%d", id)
		_ = u.redisClient.Del(context.Background(), cacheKey).Err()
	}()

	return *repo, nil
}


func (u *RepositoryUsecase) DeleteRepository(ctx context.Context, id int) error {
	ctx, span := tracing.Tracer.Start(ctx, "DeleteRepository")
	defer span.End()

	if err := u.repoRepo.Delete(ctx, id); err != nil {
		return err
	}

	event := map[string]interface{}{
		"event": "repository.deleted",
		"id":    id,
	}
	go u.kafkaProducer.Publish(event, "repository.deleted")

	// Hapus cache tanpa blocking
	go func() {
		cacheKey := fmt.Sprintf("repository:%d", id)
		_ = u.redisClient.Del(context.Background(), cacheKey).Err()
	}()

	return nil
}


