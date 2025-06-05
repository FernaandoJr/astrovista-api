package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// RateLimiter implementa um contador simples de requisições por IP com janela deslizante
type RateLimiter struct {
	mutex    sync.Mutex
	requests map[string][]time.Time
	limit    int           // Número máximo de requisições
	window   time.Duration // Janela de tempo para contagem
}

// NewRateLimiter cria um novo rate limiter com limites específicos
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Limit é um middleware que limita requisições por IP
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtém o IP do cliente (ou você pode usar um identificador personalizado)
		ip := r.RemoteAddr

		// Verifica se o IP está dentro dos limites de taxa
		if !rl.isAllowed(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")       // Sugere tentar novamente após 60 segundos
			w.WriteHeader(http.StatusTooManyRequests) // 429 Too Many Requests
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Rate limit exceeded. Please try again later.",
			})
			return
		}

		// Se está dentro dos limites, processa a requisição
		next.ServeHTTP(w, r)
	})
}

// isAllowed verifica se o IP pode fazer mais requisições
func (rl *RateLimiter) isAllowed(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Remove timestamps antigos fora da janela de tempo
	if timestamps, exists := rl.requests[ip]; exists {
		var validTimestamps []time.Time
		for _, ts := range timestamps {
			if ts.After(windowStart) {
				validTimestamps = append(validTimestamps, ts)
			}
		}
		rl.requests[ip] = validTimestamps

		// Verifica se já atingiu o limite
		if len(validTimestamps) >= rl.limit {
			return false
		}
	}

	// Adiciona a nova requisição ao contador
	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}
