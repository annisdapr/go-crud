package http

import (
	"encoding/json"
	"go-crud/internal/entity"
	"go-crud/internal/usecase"
	"net/http"
	"strconv"
	"fmt"

	"github.com/go-chi/chi/v5"
)

type RepositoryHandler struct {
	RepoUC usecase.IRepositoryUsecase
}

func NewRepositoryHandler(repoUC usecase.IRepositoryUsecase) *RepositoryHandler {
	return &RepositoryHandler{RepoUC: repoUC}
}

func (h *RepositoryHandler) CreateRepository(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Decode request body ke slice of Repository
	var repos []entity.Repository
	if err := json.NewDecoder(r.Body).Decode(&repos); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Insert tiap repository ke database
	for i := range repos {
		repos[i].UserID = userID // Set user_id dari URL
		err = h.RepoUC.CreateRepository(r.Context(), &repos[i])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Beri respons sukses dengan daftar repository yang dibuat
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(repos)
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

func (h *RepositoryHandler) GetRepositoriesByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	repos, err := h.RepoUC.GetRepositoriesByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(repos)
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("User dengan ID %d berhasil dihapus", id),
	})
}
