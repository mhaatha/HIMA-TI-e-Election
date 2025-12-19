package service

import (
	"context"

	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/web"
)

type CandidateService interface {
	Create(ctx context.Context, request web.CandidateCreateRequest) (web.CandidateResponse, error)
	GetCandidates(ctx context.Context, period string) ([]web.CandidateResponseWithURL, error)
	GetCandidateById(ctx context.Context, candidateId int) (web.CandidateResponseWithURL, error)
	UpdateCandidateById(ctx context.Context, candidateId int, request web.CandidateUpdateRequest) (web.CandidateResponse, error)
	DeleteCandidateById(ctx context.Context, candidateId int) error
}
