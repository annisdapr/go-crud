package cbreaker

import (
	"log"
	"time"
	"go-crud/internal/notifier"

	"github.com/sony/gobreaker"
)

// Fungsi reusable untuk membuat circuit breaker dengan konfigurasi default
func NewBreaker(name string) *gobreaker.CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: 3,
		Interval:    40 * time.Second,
		Timeout:     10 * time.Second,

		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Open jika terjadi 5 kegagalan berturut-turut
			return counts.ConsecutiveFailures >= 5
		},

		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logStateChange(name, from, to)
		},
	}

	return gobreaker.NewCircuitBreaker(settings)
}


func logStateChange(name string, from, to gobreaker.State) {
	stateToStr := map[gobreaker.State]string{
		gobreaker.StateClosed:   "CLOSED",
		gobreaker.StateOpen:     "OPEN",
		gobreaker.StateHalfOpen: "HALF-OPEN",
	}

	msg := "⚡ Circuit Breaker [" + name + "] berubah dari " + stateToStr[from] + " ke " + stateToStr[to]
	log.Println(msg)

	// ✅ Kirim alert ke Telegram
	notifier := notifier.NewTelegramNotifier()
	if err := notifier.SendMessage(msg); err != nil {
		log.Printf("❌ Gagal kirim notifikasi Telegram: %v", err)
	}
}
