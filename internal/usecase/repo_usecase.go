package usecase

import (
	"context"
	"errors"
	"go-crud/internal/entity"
	"go-crud/internal/repository"
	"go-crud/internal/tracing"

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
func (u *RepositoryUsecase) CreateRepository(ctx context.Context, repo *entity.Repository) error {
	ctx, span := tracing.Tracer.Start(ctx, "CreateRepository") 
	defer span.End()
	// Cek apakah user ada sebelum buat repo
	_, err := u.userRepo.GetUserByID(ctx, repo.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	return u.repoRepo.CreateRepository(ctx, repo)
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
		span.RecordError(err) // Tambahkan error logging ke tracing
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
	ctx, span := tracing.Tracer.Start(ctx, "GetRepositoriesByID") 
	defer span.End()
	return u.repoRepo.GetRepositoryByID(ctx, id)
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

	err = u.repoRepo.Update(ctx, repo) 
	if err != nil {
		return entity.Repository{}, err
	}
	return *repo, nil
}

func (u *RepositoryUsecase) DeleteRepository(ctx context.Context, id int) error {
	ctx, span := tracing.Tracer.Start(ctx, "DeleteRepository") 
	defer span.End()
	return u.repoRepo.Delete(ctx, id) 
}
