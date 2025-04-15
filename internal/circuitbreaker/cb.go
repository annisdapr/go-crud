package cbreaker

import (
	"log"
	"time"

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

// Logging perubahan state circuit breaker
func logStateChange(name string, from, to gobreaker.State) {
	stateToStr := map[gobreaker.State]string{
		gobreaker.StateClosed:   "CLOSED",
		gobreaker.StateOpen:     "OPEN",
		gobreaker.StateHalfOpen: "HALF-OPEN",
	}

	log.Printf("⚡ Circuit Breaker [%s] berubah dari %s ke %s", name, stateToStr[from], stateToStr[to])
}
