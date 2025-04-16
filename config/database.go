package config

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// DBPool adalah pool koneksi database yang bisa digunakan di seluruh aplikasi
var DBPool *pgxpool.Pool

// InitDB menginisialisasi koneksi database
func InitDB() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Tidak bisa memuat file .env, menggunakan default environment.")
	}

	// Ambil konfigurasi dari environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL tidak ditemukan dalam environment variables")
	}

	// Konfigurasi koneksi database
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Error parsing database URL: %v", err)
	}

	config.MaxConns = 10                   // Maksimum koneksi yang diizinkan
	config.MinConns = 2                    // Minimum koneksi yang aktif
	config.HealthCheckPeriod = 1 * time.Minute // Mengecek kesehatan koneksi setiap 1 menit

	// Buat pool koneksi
	DBPool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Tidak dapat terhubung ke database: %v", err)
	}

	log.Println("✅ Database berhasil terhubung!")
}

// CloseDB untuk menutup koneksi saat aplikasi berhenti
func CloseDB() {
	if DBPool != nil {
		DBPool.Close()
		log.Println("🔌 Koneksi database ditutup")
	}
}
