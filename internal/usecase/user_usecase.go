package usecase

import (
	"context"
	"go-crud/internal/entity"
	"go-crud/internal/repository"
	"time"
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
}

// Struct untuk input user (perbaikan undefined: UserInput)
type UserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// NewUserUsecase membuat instance UserUsecase
func NewUserUsecase(userRepo repository.UserRepository) IUserUsecase {
    return &UserUsecase{UserRepo: userRepo}
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

// GetUserByID mengambil data user berdasarkan ID
func (uc *UserUsecase) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	return uc.UserRepo.GetUserByID(ctx, id)
}

// Update User (perbaikan pemanggilan repository)
func (uc *UserUsecase) UpdateUser(ctx context.Context, id int, input UserInput) (entity.User, error) {
	user, err := uc.UserRepo.GetUserByID(ctx, id) // Perbaikan: gunakan GetUserByID
	if err != nil {
		return entity.User{}, err
	}

	user.Name = input.Name
	user.Email = input.Email
	user.UpdatedAt = time.Now()

	err = uc.UserRepo.UpdateUser(ctx, user) // Perbaikan: tambahkan context
	return *user, err
}

// Delete User (perbaikan pemanggilan repository)
func (uc *UserUsecase) DeleteUser(ctx context.Context, id int) error {
	return uc.UserRepo.DeleteUser(ctx, id) // Perbaikan: gunakan DeleteUser
}
