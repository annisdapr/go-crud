package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"go-crud/config"
	"go-crud/delivery"
	"go-crud/internal/repository"
	"go-crud/internal/usecase"
	"go-crud/internal/tracing"
)


var ongoingRequests int32
var wg sync.WaitGroup
func main() {
	
	shutdownTracer := tracing.InitTracer("go-crud")
	defer shutdownTracer()

	// Inisialisasi koneksi database
	config.InitDB()
	defer config.CloseDB()

	// Inisialisasi Redis
	config.InitRedis()
	defer config.CloseRedis()

	// Inisialisasi repository
	userRepo := repository.NewUserRepository(config.DBPool)
	repoRepo := repository.NewRepositoryRepository(config.DBPool)
	codeReviewRepo := repository.NewCodeReviewRepository(config.DBPool)

	// Inisialisasi usecase
	userUC := usecase.NewUserUsecase(userRepo, config.RedisClient)
	repoUC := usecase.NewRepositoryUsecase(repoRepo, userRepo)
	codeReviewUC := usecase.NewCodeReviewUsecase(codeReviewRepo, &wg)

	// Inisialisasi router dari package `delivery`
	router := delivery.NewRouter(userUC, repoUC, codeReviewUC, config.DBPool, config.RedisClient)

	// Jalankan server
	port := "8080"
	fmt.Println("🚀 Server berjalan di port", port)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Channel untuk menangani sinyal sistem
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Jalankan server di goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Gagal menjalankan server: %v", err)
		}
	}()

	// Tunggu sinyal shutdown
	<-stop
	fmt.Println("\n🛑 Menutup server...")

	// Jika masih ada proses berjalan, tunggu hingga selesai
	if atomic.LoadInt32(&ongoingRequests) > 0 {
		fmt.Printf("⚠️  Menunggu %d proses code review selesai...\n", atomic.LoadInt32(&ongoingRequests))
	}

	// Tunggu semua goroutine selesai
	wg.Wait()

	// Buat context timeout untuk shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Gagal menutup server: %v", err)
	}

	fmt.Println("✅ Server berhasil dimatikan")
}
