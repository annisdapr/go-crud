package usecase

import (
	"context"
	"fmt"
	"go-crud/internal/entity"
	"go-crud/internal/repository"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// Gunakan variabel global untuk tracking jumlah proses code review
var ongoingRequests int32

type ICodeReviewUsecase interface {
	RunCodeReview(ctx context.Context, repoID int) error
	GetReviewLogs(ctx context.Context, repoID int) ([]entity.CodeReviewLog, error)
}

type codeReviewUsecase struct {
	repo repository.CodeReviewRepository
	wg   *sync.WaitGroup // Tambahkan WaitGroup
}

// Modifikasi constructor untuk menerima WaitGroup
func NewCodeReviewUsecase(repo repository.CodeReviewRepository, wg *sync.WaitGroup) ICodeReviewUsecase {
	return &codeReviewUsecase{
		repo: repo,
		wg:   wg,
	}
}

// Simulasi Code Review (long-running task)
func (uc *codeReviewUsecase) RunCodeReview(ctx context.Context, repoID int) error {
	// Increment counter proses berjalan
	atomic.AddInt32(&ongoingRequests, 1)
	uc.wg.Add(1) // Pakai `uc.wg`, bukan `wg`

	defer func() {
		atomic.AddInt32(&ongoingRequests, -1) // Decrement setelah selesai
		uc.wg.Done()                          // Pakai `uc.wg`, bukan `wg`
	}()

	log.Println("🚀 Memulai code review untuk repo:", repoID)

	// Simulasi proses review (delay 10 detik)
	time.Sleep(10 * time.Second)

	// Hasil review (dummy result)
	reviewLog := entity.CodeReviewLog{
		RepositoryID: repoID,
		ReviewResult: "Code review completed: No critical issues found",
	}

	err := uc.repo.InsertCodeReviewLog(ctx, &reviewLog)
	if err != nil {
		return fmt.Errorf("❌ Gagal menyimpan hasil code review: %w", err)
	}

	log.Println("✅ Code review selesai untuk repo:", repoID)
	return nil
}

func (uc *codeReviewUsecase) GetReviewLogs(ctx context.Context, repoID int) ([]entity.CodeReviewLog, error) {
	return uc.repo.GetCodeReviewLogsByRepoID(ctx, repoID)
}
