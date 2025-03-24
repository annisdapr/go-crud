package http

import (
	"encoding/json"
	"go-crud/internal/entity"
	"go-crud/internal/usecase"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type RepositoryHandler struct {
	RepoUC usecase.IRepositoryUsecase
}

func NewRepositoryHandler(repoUC usecase.IRepositoryUsecase) *RepositoryHandler {
	return &RepositoryHandler{RepoUC: repoUC}
}

// Create Repository (POST /repositories)
func (h *RepositoryHandler) CreateRepository(w http.ResponseWriter, r *http.Request) {
	var repo entity.Repository
	if err := json.NewDecoder(r.Body).Decode(&repo); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := h.RepoUC.CreateRepository(r.Context(), &repo); err != nil {
		http.Error(w, "Failed to create repository", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(repo)
}

// Get Repository by ID (GET /repositories/{id})
func (h *RepositoryHandler) GetRepositoryByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid repository ID", http.StatusBadRequest)
		return
	}

	repo, err := h.RepoUC.GetRepositoryByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(repo)
}

func (h *RepositoryHandler) UpdateRepository(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var repo usecase.RepositoryInput
	if err := json.NewDecoder(r.Body).Decode(&repo); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// ✅ Perbaikan: Tambahkan r.Context() sebagai argumen pertama
	updatedRepo, err := h.RepoUC.UpdateRepository(r.Context(), id, repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedRepo)
}



// Delete Repository (DELETE /repositories/{id})
func (h *RepositoryHandler) DeleteRepository(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid repository ID", http.StatusBadRequest)
		return
	}

	err = h.RepoUC.DeleteRepository(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to delete repository", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
