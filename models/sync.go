package models

import "time"

// SyncPayload is used for syncing transactions with the backend.
type SyncPayload struct {
	Transactions []Transaction `json:"transactions"`
	LastSyncAt   time.Time     `json:"last_sync_at"`
}
