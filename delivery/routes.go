package delivery

import (
	"context"
	deliveryHTTP "go-crud/delivery/http"
	"go-crud/internal/kafka"
	"go-crud/internal/repository"
	"go-crud/internal/usecase"
	"go-crud/internal/validator"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func NewRouter(userUC usecase.IUserUsecase, repoUC usecase.IRepositoryUsecase, codeReviewUC usecase.ICodeReviewUsecase,dbPool *pgxpool.Pool, redisClient *redis.Client, mongoClient *mongo.Client) *chi.Mux {
	r := chi.NewRouter()
// ✅ Inisialisasi validator
	validator := validator.NewValidator()
	auditRepo := repository.NewAuditLogMongoRepository(mongoClient.Database("audit_log_db"))
	broker := os.Getenv("KAFKA_BROKER")
	kafkaProducer, err := kafka.NewKafkaProducer(broker)
	if err != nil {
		log.Fatalf("❌ Failed to initialize Kafka producer: %v", err)
	}


	// ✅ Inject ke handler
	userHandler := deliveryHTTP.NewUserHandler(userUC, validator, auditRepo, *kafkaProducer)

	r.Post("/users", userHandler.CreateUser)
	r.Get("/users", userHandler.GetAllUsers)
	r.Get("/users/{id}", userHandler.GetUserByID)
	r.Put("/users/{id}", userHandler.UpdateUser)
	r.Delete("/users/{id}", userHandler.DeleteUser)

	// Repository handler
	repoHandler := deliveryHTTP.NewRepositoryHandler(repoUC, validator)
	r.Post("/users/{id}/repositories", repoHandler.CreateRepository)
	r.Get("/users/{id}/repositories", repoHandler.GetRepositoriesByUserID)
	r.Get("/repositories/{id}", repoHandler.GetRepositoryByID)
	r.Get("/repositories/", repoHandler.GetAllRepositories)
	r.Put("/repositories/{id}", repoHandler.UpdateRepository)
	r.Delete("/repositories/{id}", repoHandler.DeleteRepository)

	codeReviewHandler := deliveryHTTP.NewCodeReviewHandler(context.Background(),codeReviewUC)
	r.Post("/repositories/{id}/codereview", codeReviewHandler.StartCodeReview)
	r.Get("/repositories/{id}/codereview/logs", codeReviewHandler.GetReviewLogs)

	// Health Check Handler (Sekarang menerima dbPool & Redis)
	healthHandler := deliveryHTTP.NewHealthHandler(dbPool, redisClient)
	r.Get("/health/liveness", healthHandler.LivenessCheck)
	r.Get("/health/readiness", healthHandler.ReadinessCheck)

	r.Get("/users/{id}/audit-logs", userHandler.GetUserAuditLogs)


	return r
}

