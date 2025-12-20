package domain

import "time"

type Session struct {
	SessionId     string    `json:"session_id"`
	UserId        int       `json:"user_id"`
	CreatedAt     time.Time `json:"created_at"`
	MaxAgeSeconds int       `json:"max_age_seconds"`
}
