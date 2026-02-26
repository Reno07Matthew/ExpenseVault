package api

import (
	"encoding/json"
	"net/http"

	"expenseVault/models"
)

// StartServer starts a simple HTTP server.
func StartServer(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/sync", handleSync)

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
