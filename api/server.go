package api

import (
	"encoding/json"
	"net/http"
	"time"

	"expenseVault/models"
)

// StartServer starts a simple HTTP server with middleware.
func StartServer(addr string) error {
	mux := http.NewServeMux()

	// JWT secret for auth middleware.
	jwtSecret := []byte("your-secret-key-change-in-production")

	// Rate limiter: 10 requests per second, burst of 20.
	limiter := NewRateLimiter(10, 20, time.Second)

	// /health — public (logging only)
	mux.Handle("/health", LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})))

	// /sync — protected (logging + rate limit + JWT auth)
	syncHandler := JWTAuthMiddleware(jwtSecret)(http.HandlerFunc(handleSync))
	syncHandler = RateLimitMiddleware(limiter)(syncHandler)
	mux.Handle("/sync", LoggingMiddleware(syncHandler))

	Logger.Info("Server starting", "addr", addr)
	return http.ListenAndServe(addr, mux)
}

// handleSync handles POST /sync.
// LAB 4.1: Uses models.MarshalTransactions / UnmarshalTransactions
//
//	and pointer-based payload decoding.
func handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// LAB 4: Decode into pointer — avoids copy of large payload.
	payload := &models.SyncPayload{}
	if err := json.NewDecoder(r.Body).Decode(payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// LAB 4.1: Re-marshal received transactions to validate round-trip.
	_, marshalErr := models.MarshalTransactions(payload.Transactions)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":     "ok",
		"received":   len(payload.Transactions),
		"marshal_ok": marshalErr == nil,
	})
}
