package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go-crud/config"
	"go-crud/delivery" // Gunakan `go-crud/delivery` bukan `go-crud/delivery/http`
	"go-crud/internal/repository"
	"go-crud/internal/usecase"
)

func main() {
	// Inisialisasi koneksi database
	config.InitDB()
	defer config.CloseDB()

	config.InitRedis()
	defer config.CloseRedis()

	// Inisialisasi repository
	userRepo := repository.NewUserRepository(config.DBPool)
	repoRepo := repository.NewRepositoryRepository(config.DBPool)

	// Inisialisasi usecase
	userUC := usecase.NewUserUsecase(userRepo)
	repoUC := usecase.NewRepositoryUsecase(repoRepo, userRepo)

	// Inisialisasi router dari package `delivery` // Sekarang bertipe IUserUsecase
	router := delivery.NewRouter(userUC, repoUC)
	// Mulai server
	port := "8080"
	fmt.Println("🚀 Server berjalan di port", port)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Jalankan server di goroutine agar tidak blocking
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Gagal menjalankan server: %v", err)
		}
	}()

	// Handle shutdown dengan graceful exit
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("🛑 Menutup server...")
	server.Close()
}
