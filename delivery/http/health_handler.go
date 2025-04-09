package http

import (
	"encoding/json"
	"go-crud/internal/tracing"
	"net/http"
	"sync/atomic"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
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
	ctx := r.Context()
	ctx, span := tracing.Tracer.Start(ctx, "LivenessCheck")
	defer span.End()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}


func (h *HealthHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracing.Tracer.Start(ctx, "ReadinessCheck")
	defer span.End()

	// Cek koneksi ke database
	if err := h.db.Ping(ctx); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("component", "database"), attribute.String("status", "unreachable"))
		http.Error(w, "Service not ready: DB unreachable", http.StatusServiceUnavailable)
		return
	}
	span.SetAttributes(attribute.String("db.status", "ok"))

	// Cek koneksi ke Redis
	if err := h.redis.Ping(ctx).Err(); err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("component", "redis"), attribute.String("status", "unreachable"))
		http.Error(w, "Service not ready: Redis unreachable", http.StatusServiceUnavailable)
		return
	}
	span.SetAttributes(attribute.String("redis.status", "ok"))

	span.SetAttributes(attribute.String("service.readiness", "ready"))

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
