package usecase

import (
	"context"
	"go-crud/internal/entity"
	"go-crud/internal/repository"
	"time"
	"fmt"
	"github.com/redis/go-redis/v9"
	"encoding/json" 
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
}

// Struct untuk input user (perbaikan undefined: UserInput)
type UserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// NewUserUsecase membuat instance UserUsecase
func NewUserUsecase(userRepo repository.UserRepository, redisClient *redis.Client) IUserUsecase {
    return &UserUsecase{
		UserRepo: userRepo,
		redisClient: redisClient,
	}
}


// GetAllUsers mengambil semua data user dari database
func (uc *UserUsecase) GetAllUsers(ctx context.Context) ([]entity.User, error) {
	return uc.UserRepo.GetAllUsers(ctx)
}

// CreateUser menambahkan user baru ke database
func (uc *UserUsecase) CreateUser(ctx context.Context, user *entity.User) error {
	// Atur waktu CreatedAt dan UpdatedAt
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Simpan ke database
	return uc.UserRepo.CreateUser(ctx, user)
}

// GetUserByID dengan Redis caching
func (uc *UserUsecase) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	// Buat cache key berdasarkan ID user
	cacheKey := fmt.Sprintf("user:%d", id)

	// Cek apakah user sudah ada di Redis
	val, err := uc.redisClient.Get(ctx, cacheKey).Result()
	if err == nil { // Jika ada di Redis, decode JSON ke struct User
		var cachedUser entity.User
		if err := json.Unmarshal([]byte(val), &cachedUser); err == nil {
			return &cachedUser, nil
		}
	}

	// Jika tidak ada di Redis, ambil dari database
	user, err := uc.UserRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Simpan ke Redis dengan TTL 
	userJSON, _ := json.Marshal(user)
	uc.redisClient.Set(ctx, cacheKey, userJSON, 2*time.Minute)

	return user, nil
}

// Hapus cache di Redis setelah update
func (uc *UserUsecase) UpdateUser(ctx context.Context, id int, input UserInput) (entity.User, error) {
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

	// Hapus cache di Redis setelah update
	cacheKey := fmt.Sprintf("user:%d", id)
	uc.redisClient.Del(ctx, cacheKey)

	return *user, nil
}

// Hapus cache di Redis setelah delete
func (uc *UserUsecase) DeleteUser(ctx context.Context, id int) error {
	err := uc.UserRepo.DeleteUser(ctx, id)
	if err != nil {
		return err
	}

	// Hapus cache di Redis setelah delete
	cacheKey := fmt.Sprintf("user:%d", id)
	uc.redisClient.Del(ctx, cacheKey)

	return nil
}



// // GetUserByID mengambil data user berdasarkan ID

// func (uc *UserUsecase) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
// 	return uc.UserRepo.GetUserByID(ctx, id)
// }

// // Update User (perbaikan pemanggilan repository)
// func (uc *UserUsecase) UpdateUser(ctx context.Context, id int, input UserInput) (entity.User, error) {
// 	user, err := uc.UserRepo.GetUserByID(ctx, id) // Perbaikan: gunakan GetUserByID
// 	if err != nil {
// 		return entity.User{}, err
// 	}

// 	user.Name = input.Name
// 	user.Email = input.Email
// 	user.UpdatedAt = time.Now()

// 	err = uc.UserRepo.UpdateUser(ctx, user) // Perbaikan: tambahkan context
// 	return *user, err
// }

// // Delete User (perbaikan pemanggilan repository)
// func (uc *UserUsecase) DeleteUser(ctx context.Context, id int) error {
// 	return uc.UserRepo.DeleteUser(ctx, id) // Perbaikan: gunakan DeleteUser
// }
