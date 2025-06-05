package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/HIMA-TI-e-Election/model/domain"
)

func NewVoteRepository() VoteRepository {
	return &VoteRepositoryImpl{}
}

type VoteRepositoryImpl struct{}

func (repository *VoteRepositoryImpl) GetByCandidateId(ctx context.Context, tx pgx.Tx, candidateId int) ([]domain.Vote, error) {
	SQL := `
	SELECT id, candidate_id, hashed_nim, created_at
	FROM votes
	WHERE candidate_id = $1
	`

	rows, err := tx.Query(ctx, SQL, candidateId)
	if err != nil {
		return []domain.Vote{}, nil
	}
	defer rows.Close()

	var votes []domain.Vote
	for rows.Next() {
		var vote domain.Vote

		err := rows.Scan(
			&vote.Id,
			&vote.CandidateId,
			&vote.HashedNim,
			&vote.CreatedAt,
		)

		if err != nil {
			return []domain.Vote{}, err
		}

		votes = append(votes, vote)
	}

	return votes, nil
}

func (repository *VoteRepositoryImpl) IsUserVotedInPeriod(ctx context.Context, tx pgx.Tx, hashedNIM string) (bool, error) {
	var exists bool

	SQL := `
	SELECT EXISTS (
		SELECT 1
		FROM votes v
		JOIN candidates c ON v.candidate_id = c.id
		WHERE v.hashed_nim = $1 AND EXTRACT(YEAR FROM c.created_at) = $2
	)
	`

	currentYear := time.Now().Year()

	err := tx.QueryRow(ctx, SQL, hashedNIM, currentYear).Scan(&exists)
	if err != nil {
		return exists, err
	}

	return exists, nil
}

func (repository *VoteRepositoryImpl) SaveVoteRecord(ctx context.Context, tx pgx.Tx, vote domain.Vote) (domain.Vote, error) {
	SQL := `
	INSERT INTO votes (candidate_id, hashed_nim)
	VALUES ($1, $2)
	RETURNING id, created_at
	`

	err := tx.QueryRow(ctx, SQL, vote.CandidateId, vote.HashedNim).Scan(&vote.Id, &vote.CreatedAt)
	if err != nil {
		return domain.Vote{}, err
	}

	return vote, nil
}

func (repository *VoteRepositoryImpl) GetTotalVotesByCandidateId(ctx context.Context, tx pgx.Tx, candidateId int) (int, error) {
	SQL := `
	SELECT total FROM votes_summary
	WHERE candidate_id = $1
	`

	var total int

	err := tx.QueryRow(ctx, SQL, candidateId).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}
