package domain

import "time"

type User struct {
	Id           int       `json:"id"`
	NIM          string    `json:"nim"`
	FullName     string    `json:"full_name"`
	StudyProgram string    `json:"study_program"`
	Password     string    `json:"password,omitempty"`
	Role         string    `json:"role"`
	PhoneNumber  string    `json:"phone_number"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
