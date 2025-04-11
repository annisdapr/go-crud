package http

import (
	"encoding/json"
	"go-crud/internal/entity"
	"go-crud/internal/tracing"
	"go-crud/internal/usecase"
	"go-crud/internal/validator"
	"net/http"
	"strconv"

	"fmt"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
)

type UserHandler struct {
	UserUC usecase.IUserUsecase
	Validator *validator.CustomValidator
}

func NewUserHandler(userUC usecase.IUserUsecase, validator *validator.CustomValidator) *UserHandler {
	return &UserHandler{
		UserUC: userUC,
		Validator: validator,
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

	// Tambahkan atribut dari payload ke tracing
	span.SetAttributes(
		attribute.String("user.name", user.Name),
		attribute.String("user.email", user.Email),
	)

	if err := h.UserUC.CreateUser(ctx, &user); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error.reason", "failed to create user in usecase"))
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	span.AddEvent("User successfully created")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
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
		span.SetAttributes(attribute.String("user.id_param", idStr))
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var user usecase.UserInput
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	if err := h.Validator.Validate(&user); err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Tambahkan atribut informasi dari payload
	span.SetAttributes(
		attribute.Int("user.id", id),
		attribute.String("user.name", user.Name),   // Ganti dengan field yang tersedia
		attribute.String("user.email", user.Email), // Ganti dengan field yang tersedia
	)

	updatedUser, err := h.UserUC.UpdateUser(ctx, id, user)
	if err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedUser)
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
		span.SetAttributes(attribute.String("user.id_param", idStr))
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.Int("user.id", id))

	err = h.UserUC.DeleteUser(ctx, id)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.String("delete.status", "success"))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("User dengan ID %d berhasil dihapus", id),
	})
}

// func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
// 	var user entity.User
// 	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
// 		http.Error(w, "Invalid request payload", http.StatusBadRequest)
// 		return
// 	}

	
// 	if err := h.UserUC.CreateUser(r.Context(), &user); err != nil {
// 		log.Printf("Error creating user: %v", err) 
// 		http.Error(w, "Failed to create user", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(user)
// }

