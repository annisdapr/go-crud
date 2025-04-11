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

	"github.com/joho/godotenv"

	"go-crud/config"
	"go-crud/delivery"
	"go-crud/internal/kafka"
	"go-crud/internal/repository"
	"go-crud/internal/tracing"
	"go-crud/internal/usecase"
)

var ongoingRequests int32
var wg sync.WaitGroup

func main() {
	// Load .env file (penting kalau jalan secara lokal di luar Docker)
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  .env file not found, relying on system environment variables")
	}

	// Init tracer
	shutdownTracer := tracing.InitTracer("go-crud")
	defer shutdownTracer()

	// Init PostgreSQL
	config.InitDB()
	defer config.CloseDB()

	// Init Redis
	config.InitRedis()
	defer config.CloseRedis()

	kafkaProducer, _ := kafka.NewKafkaProducer("localhost:9092", "user-events")

	// Ambil URI dan DB name dari env
	mongoURI := os.Getenv("MONGO_URI")
	mongoDBName := os.Getenv("MONGO_DB_NAME")

	if mongoURI == "" || mongoDBName == "" {
		log.Fatal("❌ MONGO_URI or MONGO_DB_NAME not set in environment variables")
	}

	// Init MongoDB
	mongoDB, mongoCleanup, err := config.InitMongoDB(mongoURI, mongoDBName)
	if err != nil {
		log.Fatal("❌ Failed to connect to MongoDB:", err)
	}
	defer mongoCleanup()

	// Inisialisasi repository dan usecase
	_ = repository.NewAuditLogMongoRepository(mongoDB)
	userRepo := repository.NewUserRepository(config.DBPool)
	repoRepo := repository.NewRepositoryRepository(config.DBPool)
	codeReviewRepo := repository.NewCodeReviewRepository(config.DBPool)

	userUC := usecase.NewUserUsecase(userRepo, config.RedisClient, kafkaProducer)
	repoUC := usecase.NewRepositoryUsecase(repoRepo, userRepo)
	codeReviewUC := usecase.NewCodeReviewUsecase(codeReviewRepo, &wg)

	// Inisialisasi router
	router := delivery.NewRouter(userUC, repoUC, codeReviewUC, config.DBPool, config.RedisClient)

	// Jalankan server HTTP
	port := "8080"
	fmt.Println("🚀 Server berjalan di port", port)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Graceful shutdown setup
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Gagal menjalankan server: %v", err)
		}
	}()

	<-stop
	fmt.Println("\n🛑 Menutup server...")

	// Tunggu semua code review selesai
	for atomic.LoadInt32(&ongoingRequests) > 0 {
		fmt.Printf("⚠️  Menunggu %d proses code review selesai...\n", atomic.LoadInt32(&ongoingRequests))
		time.Sleep(1 * time.Second)
	}

	wg.Wait()
	fmt.Println("✅ Semua code review selesai, melanjutkan shutdown server...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("❌ Gagal menutup server: %v", err)
	}

	fmt.Println("✅ Server berhasil dimatikan")
}
