package domain

import "time"

type Vote struct {
	Id          string    `json:"id"`
	CandidateId int       `json:"candidate_id"`
	HashedNim   string    `json:"hashed_nim"`
	CreatedAt   time.Time `json:"created_at"`
}
