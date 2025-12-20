package helper

import (
	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/domain"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/web"
)

func ToVoteResponse(vote domain.Vote) web.VoteResponse {
	return web.VoteResponse{
		Id:          vote.Id,
		CandidateId: vote.CandidateId,
		HashedNim:   vote.HashedNim,
		CreatedAt:   vote.CreatedAt,
	}
}

func ToVotesResponse(votes []domain.Vote) []web.VoteResponse {
	var voteResponses []web.VoteResponse
	for _, vote := range votes {
		voteResponses = append(voteResponses, ToVoteResponse(vote))
	}
	return voteResponses
}
