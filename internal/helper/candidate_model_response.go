package helper

import (
	"strings"

	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/domain"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/web"
)

func ToCandidateResponse(candidate domain.Candidate) web.CandidateResponse {
	mission := strings.Split(candidate.Mission, "*")

	return web.CandidateResponse{
		Id:                    candidate.Id,
		Number:                candidate.Number,
		President:             candidate.President,
		Vice:                  candidate.Vice,
		Vision:                candidate.Vision,
		Mission:               mission,
		PhotoKey:              candidate.PhotoKey,
		PresidentStudyProgram: candidate.PresidentStudyProgram,
		ViceStudyProgram:      candidate.ViceStudyProgram,
		PresidentNIM:          candidate.PresidentNIM,
		ViceNIM:               candidate.ViceNIM,
		CreatedAt:             candidate.CreatedAt,
		UpdatedAt:             candidate.UpdatedAt,
	}
}

func ToCandidatesResponse(candidates []domain.Candidate) []web.CandidateResponse {
	var candidateResponses []web.CandidateResponse
	for _, candidate := range candidates {
		candidateResponses = append(candidateResponses, ToCandidateResponse(candidate))
	}
	return candidateResponses
}

func ToCandidateResponseWithURL(candidate domain.CandidateWithURL) web.CandidateResponseWithURL {
	mission := strings.Split(candidate.Mission, "*")

	return web.CandidateResponseWithURL{
		Id:                    candidate.Id,
		Number:                candidate.Number,
		President:             candidate.President,
		Vice:                  candidate.Vice,
		Vision:                candidate.Vision,
		Mission:               mission,
		PhotoURL:              candidate.PhotoURL,
		PresidentStudyProgram: candidate.PresidentStudyProgram,
		ViceStudyProgram:      candidate.ViceStudyProgram,
		PresidentNIM:          candidate.PresidentNIM,
		ViceNIM:               candidate.ViceNIM,
		CreatedAt:             candidate.CreatedAt,
		UpdatedAt:             candidate.UpdatedAt,
	}
}

func ToCandidatesResponseWithURL(candidates []domain.CandidateWithURL) []web.CandidateResponseWithURL {
	var candidateResponsesWithURL []web.CandidateResponseWithURL
	for _, candidate := range candidates {
		candidateResponsesWithURL = append(candidateResponsesWithURL, ToCandidateResponseWithURL(candidate))
	}
	return candidateResponsesWithURL
}
