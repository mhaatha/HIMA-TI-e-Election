package repository

import (
	"context"
	"database/sql"

	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/domain"
)

type VoteRepository interface {
	GetByCandidateId(ctx context.Context, tx *sql.Tx, candidateId int) ([]domain.Vote, error)
	IsUserVotedInPeriod(ctx context.Context, tx *sql.Tx, hashedNIM string) (bool, error)
	SaveVoteRecord(ctx context.Context, tx *sql.Tx, vote domain.Vote) (domain.Vote, error)
	GetTotalVotesByCandidateId(ctx context.Context, tx *sql.Tx, candidateId int) (int, error)
	IsUserEverVoted(ctx context.Context, tx *sql.Tx, userId int) (bool, error)
}
