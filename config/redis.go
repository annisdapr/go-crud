package config

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/joho/godotenv"
)

// RedisClient adalah instance Redis yang bisa digunakan di seluruh aplikasi
var RedisClient *redis.Client

// InitRedis menginisialisasi koneksi Redis
func InitRedis() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Tidak bisa memuat file .env, menggunakan default environment.")
	}

	// Ambil konfigurasi dari environment variables
	redisAddr := os.Getenv("REDIS_ADDR")           // Contoh: localhost:6379
	redisPassword := os.Getenv("REDIS_PASSWORD")   // Jika kosong, Redis tidak pakai password
	redisDB, _ := strconv.Atoi(os.Getenv("REDIS_DB")) // Redis database index

	// Inisialisasi koneksi Redis
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword, // Kosongkan jika tidak ada password
		DB:       redisDB,
	})

	// Cek koneksi Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("❌ Tidak dapat terhubung ke Redis: %v", err)
	}

	log.Println("✅ Redis berhasil terhubung di", redisAddr)
}

// CloseRedis untuk menutup koneksi Redis saat aplikasi berhenti
func CloseRedis() {
	if RedisClient != nil {
		RedisClient.Close()
		log.Println("🔌 Koneksi Redis ditutup")
	}
}
