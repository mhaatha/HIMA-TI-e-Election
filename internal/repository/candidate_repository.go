package repository

import (
	"context"
	"database/sql"

	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/domain"
)

type CandidateRepository interface {
	Save(ctx context.Context, tx *sql.Tx, candidate domain.Candidate) (domain.Candidate, error)
	GetAll(ctx context.Context, tx *sql.Tx) ([]domain.Candidate, error)
	GetByPeriod(ctx context.Context, tx *sql.Tx, period int) ([]domain.Candidate, error)
	GetById(ctx context.Context, tx *sql.Tx, candidateId int) (domain.Candidate, error)
	UpdateById(ctx context.Context, tx *sql.Tx, candidateId int, candidate domain.Candidate) (domain.Candidate, error)
	DeleteById(ctx context.Context, tx *sql.Tx, candidateId int) error
}
