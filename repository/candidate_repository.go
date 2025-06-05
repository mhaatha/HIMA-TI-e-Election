package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/HIMA-TI-e-Election/model/domain"
)

type CandidateRepository interface {
	Save(ctx context.Context, tx pgx.Tx, candidate domain.Candidate) (domain.Candidate, error)
	GetAll(ctx context.Context, tx pgx.Tx) ([]domain.Candidate, error)
	GetByPeriod(ctx context.Context, tx pgx.Tx, period int) ([]domain.Candidate, error)
	GetById(ctx context.Context, tx pgx.Tx, candidateId int) (domain.Candidate, error)
	UpdateById(ctx context.Context, tx pgx.Tx, candidateId int, candidate domain.Candidate) (domain.Candidate, error)
	DeleteById(ctx context.Context, tx pgx.Tx, candidateId int) error
}
