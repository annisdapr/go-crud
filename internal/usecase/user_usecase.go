package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"go-crud/internal/entity"
	"go-crud/internal/kafka"
	"go-crud/internal/repository"
	"go-crud/internal/tracing"

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
}

// UserUsecase mengelola logika bisnis untuk User
type UserUsecase struct {
    UserRepo repository.UserRepository
	redisClient *redis.Client
	kafkaProducer *kafka.KafkaProducer 
	cbRedis        *gobreaker.CircuitBreaker
	cbPostgres *gobreaker.CircuitBreaker

}

// Struct untuk input user (perbaikan undefined: UserInput)
type UserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// NewUserUsecase membuat instance UserUsecase
func NewUserUsecase(userRepo repository.UserRepository, redisClient *redis.Client, kafkaProducer *kafka.KafkaProducer) IUserUsecase {
	cbRedis := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:    "RedisBreaker",
		Timeout: 5 * time.Second,
	})

	cbPostgres := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:    "PostgresBreaker",
		Timeout: 5 * time.Second,
	})
	return &UserUsecase{
		UserRepo:      userRepo,
		redisClient:   redisClient,
		kafkaProducer: kafkaProducer,
		cbRedis:       cbRedis,
		cbPostgres:    cbPostgres,
	}
}


// GetAllUsers mengambil semua data user dari database
func (uc *UserUsecase) GetAllUsers(ctx context.Context) ([]entity.User, error) {
	ctx, span := tracing.Tracer.Start(ctx, "UserUsecase.GetAllUsers") 
	defer span.End()

	return uc.UserRepo.GetAllUsers(ctx)
}

// CreateUser mengirim event ke Kafka untuk dibuat oleh consumer
func (uc *UserUsecase) CreateUser(ctx context.Context, user *entity.User) error {
	ctx, span := tracing.Tracer.Start(ctx, "UserUsecase.CreateUser")
	defer span.End()

	// Set waktu sekarang
	now := time.Now()

	// Siapkan event payload
	event := map[string]interface{}{
		"event":      "user.created",
		"id":         user.ID, 
		"name":       user.Name,
		"email":      user.Email,
		"time":       now.Format(time.RFC3339),
	}

	// Publish ke Kafka
	go uc.kafkaProducer.Publish(event, "user.created")

	// Tidak langsung insert ke database
	return nil
}



// GetUserByID dengan Redis caching
func (uc *UserUsecase) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	ctx, span := tracing.Tracer.Start(ctx, "UserUsecase.GetUserById")
	defer span.End()

	cacheKey := fmt.Sprintf("user:%d", id)

	// ✅ Ambil dari Redis dengan circuit breaker
	val, err := uc.cbRedis.Execute(func() (interface{}, error) {
		return uc.redisClient.Get(ctx, cacheKey).Result()
	})

	if err == nil {
		var cachedUser entity.User
		if unmarshalErr := json.Unmarshal([]byte(val.(string)), &cachedUser); unmarshalErr == nil {
			return &cachedUser, nil
		}
	}

	// ✅ Ambil dari PostgreSQL dengan circuit breaker
	result, err := uc.cbPostgres.Execute(func() (interface{}, error) {
		return uc.UserRepo.GetUserByID(ctx, id)
	})
	if err != nil {
		return nil, err
	}

	user := result.(*entity.User)

	// ✅ Simpan ke Redis dengan circuit breaker
	userJSON, _ := json.Marshal(user)
	_, _ = uc.cbRedis.Execute(func() (interface{}, error) {
		return nil, uc.redisClient.Set(ctx, cacheKey, userJSON, 2*time.Minute).Err()
	})

	return user, nil
}


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

	cacheKey := fmt.Sprintf("user:%d", id)
	uc.redisClient.Del(ctx, cacheKey)

	// 👇 Publish Kafka event
	event := map[string]interface{}{
		"event": "user.updated",
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"time":  user.UpdatedAt.Format(time.RFC3339),
	}

	go uc.kafkaProducer.Publish(event, "user.updated")

	return *user, nil
}
// Hapus cache di Redis setelah delete
// DeleteUser mengirim event user.deleted ke Kafka
func (uc *UserUsecase) DeleteUser(ctx context.Context, id int) error {
	ctx, span := tracing.Tracer.Start(ctx, "UserUsecase.DeleteUser")
	defer span.End()

	// Buat event payload
	event := map[string]interface{}{
		"event": "user.deleted",
		"id":    id,
		"time":  time.Now().Format(time.RFC3339),
	}

	// Publish ke Kafka
	go uc.kafkaProducer.Publish(event, "user.deleted")

	// Tidak langsung hapus dari database
	return nil
}
