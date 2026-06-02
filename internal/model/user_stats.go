package model

import "time"

type UserStats struct {
	UserID        string
	ArchivedCount int
	UpdatedAt     time.Time
}
