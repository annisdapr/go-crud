package usecase

import (
	"context"
	"database/sql"
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


type IUserUsecase interface {
    CreateUser(ctx context.Context, user *entity.User) error
    GetUserByID(ctx context.Context, id int) (*entity.User, error)
    UpdateUser(ctx context.Context, id int, input UserInput) (entity.User, error)
    DeleteUser(ctx context.Context, id int) error
	GetAllUsers(ctx context.Context) ([]entity.User, error) 
	IsEmailExists(ctx context.Context, email string) (bool, error)
}

// UserUsecase mengelola logika bisnis untuk User
type UserUsecase struct {
	UserRepo   repository.UserRepository
	redis      *redis.Client
	cbRedis    *gobreaker.CircuitBreaker
	cbPostgres *gobreaker.CircuitBreaker
}

type UserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// NewUserUsecase membuat instance UserUsecase

func NewUserUsecase(userRepo repository.UserRepository, redisClient *redis.Client) IUserUsecase {
	return &UserUsecase{
		UserRepo:   userRepo,
		redis:      redisClient,
		cbRedis:    cbreaker.NewBreaker("RedisBreaker"),
		cbPostgres: cbreaker.NewBreaker("PostgresBreaker"),
	}
}

func (uc *UserUsecase) IsEmailExists(ctx context.Context, email string) (bool, error) {
	ctx, span := tracing.Tracer.Start(ctx, "UserUsecase.IsEmailExists")
	defer span.End()

	_, err := uc.UserRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ✅ Validasi untuk user creation
func (uc *UserUsecase) CreateUser(ctx context.Context, user *entity.User) error {
	ctx, span := tracing.Tracer.Start(ctx, "UserUsecase.CreateUser")
	defer span.End()

	// Validasi duplicate email
	exists, err := uc.IsEmailExists(ctx, user.Email)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("email already in use")
	}

	// Tidak insert ke DB langsung (tugas Kafka consumer)
	return nil
}

// ✅ Get all users langsung ke repo
func (uc *UserUsecase) GetAllUsers(ctx context.Context) ([]entity.User, error) {
	ctx, span := tracing.Tracer.Start(ctx, "UserUsecase.GetAllUsers")
	defer span.End()

	return uc.UserRepo.GetAllUsers(ctx)
}

// ✅ Get user dari cache (Redis) atau DB
func (uc *UserUsecase) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	ctx, span := tracing.Tracer.Start(ctx, "UserUsecase.GetUserById")
	defer span.End()

	cacheKey := fmt.Sprintf("user:%d", id)

	// Coba ambil dari Redis
	val, err := uc.cbRedis.Execute(func() (interface{}, error) {
		return uc.redis.Get(ctx, cacheKey).Result()
	})
	if err == nil {
		var cachedUser entity.User
		if unmarshalErr := json.Unmarshal([]byte(val.(string)), &cachedUser); unmarshalErr == nil {
			return &cachedUser, nil
		}
	}

	// Ambil dari DB
	result, err := uc.cbPostgres.Execute(func() (interface{}, error) {
		return uc.UserRepo.GetUserByID(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	return result.(*entity.User), nil
}

// ✅ Update user (hanya update DB dan return hasil)
func (uc *UserUsecase) UpdateUser(ctx context.Context, id int, input UserInput) (entity.User, error) {
	ctx, span := tracing.Tracer.Start(ctx, "UserUsecase.UpdateUser")
	defer span.End()

	user, err := uc.UserRepo.GetUserByID(ctx, id)
	if err != nil {
		return entity.User{}, err
	}

	user.Name = input.Name
	user.Email = input.Email
	user.UpdatedAt = time.Now()

	err = uc.UserRepo.UpdateUser(ctx, user)
	if err != nil {
		return entity.User{}, err
	}

	// Cache dan audit dilakukan di layer atas
	return *user, nil
}

// ✅ Delete user (tugas Kafka consumer nanti)
func (uc *UserUsecase) DeleteUser(ctx context.Context, id int) error {
	_, span := tracing.Tracer.Start(ctx, "UserUsecase.DeleteUser")
	defer span.End()

	// Tidak hapus langsung dari DB
	return nil
}
