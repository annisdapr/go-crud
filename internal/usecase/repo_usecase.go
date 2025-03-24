package usecase

import (
	"context"
	"errors"
	"go-crud/internal/entity"
	"go-crud/internal/repository"
)

// Interface untuk RepositoryUsecase
type IRepositoryUsecase interface {
	CreateRepository(ctx context.Context, repo *entity.Repository) error
	GetRepositoryByID(ctx context.Context, id int) (*entity.Repository, error)
	UpdateRepository(ctx context.Context, id int, input RepositoryInput) (entity.Repository, error)
	DeleteRepository(ctx context.Context, id int) error
}

// Struct RepositoryUsecase
type RepositoryUsecase struct {
	repoRepo repository.RepositoryRepository
	userRepo repository.UserRepository
}

// Input struct untuk repository
type RepositoryInput struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	AIEnabled bool   `json:"ai_enabled"`
}

// NewRepositoryUsecase membuat instance baru
func NewRepositoryUsecase(repoRepo repository.RepositoryRepository, userRepo repository.UserRepository) IRepositoryUsecase {
	return &RepositoryUsecase{repoRepo: repoRepo, userRepo: userRepo}
}

// ✅ Perbaiki Implementasi Metode
func (u *RepositoryUsecase) CreateRepository(ctx context.Context, repo *entity.Repository) error {
	// Cek apakah user ada sebelum buat repo
	_, err := u.userRepo.GetUserByID(ctx, repo.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	return u.repoRepo.CreateRepository(ctx, repo)
}

func (u *RepositoryUsecase) GetRepositoryByID(ctx context.Context, id int) (*entity.Repository, error) {
	return u.repoRepo.GetRepositoryByID(ctx, id)
}

// ✅ Perbaikan Pemanggilan Repository
func (u *RepositoryUsecase) UpdateRepository(ctx context.Context, id int, input RepositoryInput) (entity.Repository, error) {
	repo, err := u.repoRepo.GetRepositoryByID(ctx, id) // ✅ Perbaiki pemanggilan method
	if err != nil {
		return entity.Repository{}, err
	}

	repo.Name = input.Name
	repo.URL = input.URL
	repo.AIEnabled = input.AIEnabled

	err = u.repoRepo.Update(ctx, repo) // ✅ Perbaiki pemanggilan method
	if err != nil {
		return entity.Repository{}, err
	}
	return *repo, nil
}

func (u *RepositoryUsecase) DeleteRepository(ctx context.Context, id int) error {
	return u.repoRepo.Delete(ctx, id) // ✅ Perbaiki pemanggilan method
}
