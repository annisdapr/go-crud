package http

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9" 
)

type HealthHandler struct {
	db    *pgxpool.Pool
	redis *redis.Client 
	ready atomic.Bool
}

// NewHealthHandler menerima database pool dan Redis client
func NewHealthHandler(db *pgxpool.Pool, redisClient *redis.Client) *HealthHandler {
	h := &HealthHandler{
		db:    db,
		redis: redisClient,
	}
	h.ready.Store(false)
	return h
}

// LivenessCheck memastikan aplikasi masih berjalan
func (h *HealthHandler) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

// ReadinessCheck memastikan aplikasi siap menerima trafik
func (h *HealthHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Cek koneksi ke database
	if err := h.db.Ping(ctx); err != nil {
		http.Error(w, "Service not ready: DB unreachable", http.StatusServiceUnavailable)
		return
	}

	// Cek koneksi ke Redis
	if err := h.redis.Ping(ctx).Err(); err != nil {
		http.Error(w, "Service not ready: Redis unreachable", http.StatusServiceUnavailable)
		return
	}

	// Jika semua OK
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

// // SetReady menandakan bahwa aplikasi siap menerima trafik
// func (h *HealthHandler) SetReady() {
// 	h.ready.Store(true)
// }

// // SetNotReady menandakan bahwa aplikasi belum siap menerima trafik
// func (h *HealthHandler) SetNotReady() {
// 	h.ready.Store(false)
// }
