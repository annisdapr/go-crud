package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-crud/config"
	"go-crud/delivery"
	deliveryHTTP "go-crud/delivery/http"
	"go-crud/internal/repository"
	"go-crud/internal/usecase"
)

func main() {
	// Inisialisasi koneksi database
	config.InitDB()
	defer config.CloseDB()

	// Inisialisasi Redis
	config.InitRedis()
	defer config.CloseRedis()

	// Inisialisasi repository
	userRepo := repository.NewUserRepository(config.DBPool)
	repoRepo := repository.NewRepositoryRepository(config.DBPool)

	// Inisialisasi usecase
	userUC := usecase.NewUserUsecase(userRepo)
	repoUC := usecase.NewRepositoryUsecase(repoRepo, userRepo)

	// Inisialisasi health handler
	healthHandler := deliveryHTTP.NewHealthHandler(config.DBPool, config.RedisClient)

	// Inisialisasi router dari package `delivery`
	router := delivery.NewRouter(userUC, repoUC, config.DBPool, config.RedisClient)

	// Tambahkan health check handler
	router.Get("/health/liveness", healthHandler.LivenessCheck)
	router.Get("/health/readiness", healthHandler.ReadinessCheck)

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

	// Jalankan server di goroutine agar tidak blocking
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Gagal menjalankan server: %v", err)
		}
	}()

	// Tunggu sinyal shutdown
	<-stop
	fmt.Println("🛑 Menutup server...")

	// Buat context timeout untuk graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Tutup server secara graceful
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Gagal menutup server: %v", err)
	}

	fmt.Println("✅ Server berhasil dimatikan")
}
