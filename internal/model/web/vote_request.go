package web

type VoteCreateRequest struct {
	CandidateId int `json:"candidate_id" validate:"required"`
}
