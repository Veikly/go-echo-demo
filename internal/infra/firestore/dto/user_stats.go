package dto

import "time"

type UserStats struct {
	UserID        string    `firestore:"user_id"`
	ArchivedCount int       `firestore:"archived_count"`
	UpdatedAt     time.Time `firestore:"updated_at"`
}
