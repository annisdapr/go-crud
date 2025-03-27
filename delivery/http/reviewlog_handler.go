package http

import (
	"encoding/json"
	"go-crud/internal/usecase"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CodeReviewHandler struct {
	CodeReviewUC usecase.ICodeReviewUsecase
}

func NewCodeReviewHandler(uc usecase.ICodeReviewUsecase) *CodeReviewHandler {
	return &CodeReviewHandler{CodeReviewUC: uc}
}

// Mulai code review (long-running task)
func (h *CodeReviewHandler) StartCodeReview(w http.ResponseWriter, r *http.Request) {
	repoID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid repository ID", http.StatusBadRequest)
		return
	}

	// Tambahkan tracking ke goroutine
	go func() {
		_ = h.CodeReviewUC.RunCodeReview(r.Context(), repoID)
	}()

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "🚀 Code review is in progress",
	})
}

// Ambil hasil code review berdasarkan repository ID
func (h *CodeReviewHandler) GetReviewLogs(w http.ResponseWriter, r *http.Request) {
	repoID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid repository ID", http.StatusBadRequest)
		return
	}

	logs, err := h.CodeReviewUC.GetReviewLogs(r.Context(), repoID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(logs)
}
