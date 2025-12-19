package service

import (
	"context"

	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/web"
)

type VoteService interface {
	// Method to solve Circular Dependency
	SetCandidateService(candidateService CandidateService)

	GetByCandidateId(ctx context.Context, candidateId int) ([]web.VoteResponse, error)
	SaveVoteRecord(ctx context.Context, request web.VoteCreateRequest, userId int) (web.VoteCreateResponse, error)
	GetTotalVotesByCandidateId(ctx context.Context, candidateId int) (web.TotalVoteResponse, error)
	CheckIfUserHasVoted(ctx context.Context, sessionId string) (bool, error)
}
