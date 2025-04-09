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
	wg   *sync.WaitGroup 
}

// Modifikasi constructor untuk menerima WaitGroup
func NewCodeReviewUsecase(repo repository.CodeReviewRepository, wg *sync.WaitGroup) ICodeReviewUsecase {
	return &codeReviewUsecase{
		repo: repo,
		wg:   wg,
	}
}

func (uc *codeReviewUsecase) RunCodeReview(ctx context.Context, repoID int) error {
	atomic.AddInt32(&ongoingRequests, 1)
	uc.wg.Add(1) 

	defer func() {
		atomic.AddInt32(&ongoingRequests, -1)
		uc.wg.Done()
	}()

	log.Println("🚀 Memulai code review untuk repo:", repoID)

	// Simulasi kerja selama 10 detik
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done(): // Jika context dibatalkan, beri kesempatan untuk menyelesaikan
			log.Println("⏳ Code review sedang dalam proses, menunggu hingga aman untuk keluar...")

			// Tunggu sebentar untuk menyelesaikan tugas dengan aman
			time.Sleep(2 * time.Second)
			log.Println("⏹️ Code review dihentikan untuk repo:", repoID)
			return ctx.Err()
		default:
			time.Sleep(1 * time.Second) // Simulasi kerja per detik
		}
	}

	// Simpan hasil review
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

// // Simulasi Code Review (long-running task)
// func (uc *codeReviewUsecase) RunCodeReview(ctx context.Context, repoID int) error {
// 	atomic.AddInt32(&ongoingRequests, 1)
// 	uc.wg.Add(1) 

// 	defer func() {
// 		atomic.AddInt32(&ongoingRequests, -1)
// 		uc.wg.Done()
// 	}()

// 	log.Println("🚀 Memulai code review untuk repo:", repoID)

// 	// Loop untuk bisa menangkap sinyal shutdown saat sleep
// 	for i := 0; i < 10; i++ {
// 		select {
// 		case <-ctx.Done(): // Jika context dibatalkan, hentikan proses
// 			log.Println("⏹️ Code review dibatalkan untuk repo:", repoID)
// 			return ctx.Err()
// 		default:
// 			time.Sleep(1 * time.Second) // Tunggu per detik untuk bisa dicek
// 		}
// 	}

// 	// Hasil review (dummy result)
// 	reviewLog := entity.CodeReviewLog{
// 		RepositoryID: repoID,
// 		ReviewResult: "Code review completed: No critical issues found",
// 	}

// 	err := uc.repo.InsertCodeReviewLog(ctx, &reviewLog)
// 	if err != nil {
// 		return fmt.Errorf("❌ Gagal menyimpan hasil code review: %w", err)
// 	}

// 	log.Println("✅ Code review selesai untuk repo:", repoID)
// 	return nil
// }


func (uc *codeReviewUsecase) GetReviewLogs(ctx context.Context, repoID int) ([]entity.CodeReviewLog, error) {
	return uc.repo.GetCodeReviewLogsByRepoID(ctx, repoID)
}
