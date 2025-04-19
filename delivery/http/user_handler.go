package http

import (
	"encoding/json"
	"go-crud/internal/entity"
	"go-crud/internal/tracing"
	"go-crud/internal/usecase"
	"go-crud/internal/validator"
	"go-crud/internal/repository"
	"go-crud/internal/kafka"
	"net/http"
	"strconv"

	"fmt"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
)

type UserHandler struct {
	UserUC usecase.IUserUsecase
	Validator *validator.CustomValidator
	AuditRepo   repository.AuditLogMongoRepository
	Producer   kafka.KafkaProducer
}

func NewUserHandler(userUC usecase.IUserUsecase, validator *validator.CustomValidator, auditRepo repository.AuditLogMongoRepository, producer kafka.KafkaProducer) *UserHandler {
	return &UserHandler{
		UserUC: userUC,
		Validator: validator,
		AuditRepo: auditRepo,
		Producer:   producer,
	}
}


// Create User (POST /users)
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracing.Tracer.Start(ctx, "UserHandler.CreateUser")
	defer span.End()

	var user entity.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.reason", "invalid JSON payload"))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := h.Validator.Validate(&user); err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ✅ Cek apakah email sudah digunakan
	exists, err := h.UserUC.IsEmailExists(ctx, user.Email)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}

	span.SetAttributes(
		attribute.String("user.name", user.Name),
		attribute.String("user.email", user.Email),
	)

	eventData := map[string]interface{}{
		"name":  user.Name,
		"email": user.Email,
	}

	// payload, err := json.Marshal(eventData)
	// if err != nil {
	// 	span.RecordError(err)
	// 	http.Error(w, "Failed to marshal user data", http.StatusInternalServerError)
	// 	return
	// }

	// ✅ Kirim ke Kafka
	err = h.Producer.Publish("user-events", eventData, "user.created")

	if err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to send event to Kafka", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Create user event sent to Kafka",
	})
}


// GetAllUsers (GET /users)
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracing.Tracer.Start(ctx, "UserHandler.GetAllUsers")
	defer span.End()

	users, err := h.UserUC.GetAllUsers(ctx)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	// Tambahkan jumlah user sebagai atribut tracing
	span.SetAttributes(
		attribute.Int("user.count", len(users)),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}


// Get User by ID (GET /users/{id})
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracing.Tracer.Start(ctx, "UserHandler.GetUserByID")
	defer span.End()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("user.id_param", idStr))
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.UserUC.GetUserByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Int("user.requested_id", id))
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Tambahkan atribut informasi user jika berhasil ditemukan
	span.SetAttributes(
		attribute.Int("user.id", user.ID),
		attribute.String("user.name", user.Name), // Ganti field jika perlu
	)

	json.NewEncoder(w).Encode(user)
}

// Update User (PATCH /users/{id})
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracing.Tracer.Start(ctx, "UserHandler.UpdateUser")
	defer span.End()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var input usecase.UserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.Validator.Validate(&input); err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.Int("user.id", id),
		attribute.String("user.name", input.Name),
		attribute.String("user.email", input.Email),
	)

	if _, err := h.UserUC.UpdateUser(ctx, id, input); err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to send update event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Update user event sent to Kafka",
	})
}


// Delete User (DELETE /users/{id})
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracing.Tracer.Start(ctx, "UserHandler.DeleteUser")
	defer span.End()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.UserUC.DeleteUser(ctx, id); err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to send delete event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Delete user event for ID %d sent to Kafka", id),
	})
}

func (h *UserHandler) GetUserAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Ambil user ID dari URL path parameter dengan chi
	idStr := chi.URLParam(r, "id") // Ambil id dari path parameter {id}

	if idStr == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	logs, err := h.AuditRepo.GetLogs(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

