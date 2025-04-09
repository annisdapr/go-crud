package http

import (
	"encoding/json"
	"fmt"
	"go-crud/internal/entity"
	"go-crud/internal/tracing"
	"go-crud/internal/usecase"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type RepositoryHandler struct {
	RepoUC usecase.IRepositoryUsecase
}

func NewRepositoryHandler(repoUC usecase.IRepositoryUsecase) *RepositoryHandler {
	return &RepositoryHandler{RepoUC: repoUC}
}

func (h *RepositoryHandler) CreateRepository(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracing.Tracer.Start(ctx, "CreateRepository")
	defer span.End()

	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	span.SetAttributes(attribute.Int("user.id", userID))

	// Decode request body ke slice of Repository
	var repos []entity.Repository
	if err := json.NewDecoder(r.Body).Decode(&repos); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	span.SetAttributes(attribute.Int("repository.requested_count", len(repos)))

	// Insert tiap repository ke database
	for i := range repos {
		repos[i].UserID = userID
		err = h.RepoUC.CreateRepository(ctx, &repos[i])
		if err != nil {
			span.RecordError(err)
			span.AddEvent("Failed to create one of the repositories", trace.WithAttributes(
				attribute.String("repository.name", repos[i].Name),
			))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	span.AddEvent("All repositories successfully created")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(repos)
}



func (h *RepositoryHandler) GetAllRepositories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracing.Tracer.Start(ctx, "GetAllRepositories")
	defer span.End()

	repos, err := h.RepoUC.GetAllRepositories(ctx)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to fetch repositories", http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Int("repository.count", len(repos)))
	span.AddEvent("Successfully fetched all repositories")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repos)
}


// Get Repository by ID (GET /repositories/{id})
func (h *RepositoryHandler) GetRepositoryByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() 
	ctx, span := tracing.Tracer.Start(ctx, "GetRepositoryByID")
	defer span.End()
	
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		span.RecordError(err) 
		http.Error(w, "Invalid repository ID", http.StatusBadRequest)
		return
	}

	repo, err := h.RepoUC.GetRepositoryByID(ctx, id) 
	if err != nil {
		span.RecordError(err) 
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(repo)
}

func (h *RepositoryHandler) GetRepositoriesByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() 
	ctx, span := tracing.Tracer.Start(ctx, "GetRepositoryByUserID")
	defer span.End()

	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.Int("user.id", userID))

	repos, err := h.RepoUC.GetRepositoriesByUserID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(repos)
}



func (h *RepositoryHandler) UpdateRepository(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracing.Tracer.Start(ctx, "UpdateRepository")
	defer span.End()

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	span.SetAttributes(attribute.Int("repository.id", id))

	var repo usecase.RepositoryInput
	if err := json.NewDecoder(r.Body).Decode(&repo); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("repository.name", repo.Name),
		attribute.String("repository.url", repo.URL),
		attribute.Bool("repository.ai_enabled", repo.AIEnabled),
	)

	updatedRepo, err := h.RepoUC.UpdateRepository(ctx, id, repo)
	if err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedRepo)
}



func (h *RepositoryHandler) DeleteRepository(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracing.Tracer.Start(ctx, "DeleteRepository")
	defer span.End()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid repository ID", http.StatusBadRequest)
		return
	}
	span.SetAttributes(attribute.Int("repository.id", id))

	err = h.RepoUC.DeleteRepository(ctx, id)
	if err != nil {
		span.RecordError(err)
		http.Error(w, "Failed to delete repository", http.StatusInternalServerError)
		return
	}

	span.AddEvent("Repository successfully deleted")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Repository dengan ID %d berhasil dihapus", id),
	})
}

// Create Repository (POST /repositories)
// func (h *RepositoryHandler) CreateRepository(w http.ResponseWriter, r *http.Request) {
// 	var repo entity.Repository
// 	if err := json.NewDecoder(r.Body).Decode(&repo); err != nil {
// 		http.Error(w, "Invalid request payload", http.StatusBadRequest)
// 		return
// 	}

// 	if err := h.RepoUC.CreateRepository(r.Context(), &repo); err != nil {
// 		http.Error(w, "Failed to create repository", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(repo)
// }