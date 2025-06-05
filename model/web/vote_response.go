package web

import "time"

type VoteResponse struct {
	Id          string    `json:"id"`
	CandidateId int       `json:"candidate_id"`
	HashedNim   string    `json:"hashed_nim"`
	CreatedAt   time.Time `json:"created_at"`
}

type VoteCreateResponse struct {
	CreatedAt time.Time `json:"created_at"`
}

type TotalVoteResponse struct {
	TotalVotes int `json:"total_votes"`
}
