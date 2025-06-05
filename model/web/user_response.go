package web

import "time"

type UserResponse struct {
	ID           int       `json:"id"`
	NIM          string    `json:"nim"`
	FullName     string    `json:"full_name"`
	StudyProgram string    `json:"study_program"`
	Role         string    `json:"role"`
	PhoneNumber  string    `json:"phone_number"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserGetByNimResponse struct {
	ID           int       `json:"id"`
	NIM          string    `json:"nim"`
	FullName     string    `json:"full_name"`
	StudyProgram string    `json:"study_program"`
	Password     string    `json:"password,omitempty"`
	Role         string    `json:"role"`
	PhoneNumber  string    `json:"phone_number"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SessionResponse struct {
	SessionId     string    `json:"session_id"`
	UserId        int       `json:"user_id"`
	CreatedAt     time.Time `json:"created_at"`
	MaxAgeSeconds int       `json:"max_age_seconds"`
}

type AdminResponse struct {
	ID           int       `json:"id"`
	NIM          string    `json:"nim"`
	FullName     string    `json:"full_name"`
	StudyProgram string    `json:"study_program"`
	Password     string    `json:"password"`
	Role         string    `json:"role"`
	PhoneNumber  string    `json:"phone_number"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
