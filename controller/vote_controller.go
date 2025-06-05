package controller

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type VoteController interface {
	Save(w http.ResponseWriter, r *http.Request, params httprouter.Params)
	GetTotalVotesByCandidateId(w http.ResponseWriter, r *http.Request, params httprouter.Params)
	VotesLiveResult(w http.ResponseWriter, r *http.Request, params httprouter.Params)
	ListenToDB(ctx context.Context)
}
