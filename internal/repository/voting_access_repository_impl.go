package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/domain"
)

func NewVotingAccessRepository() VotingAccessRepository {
	return &VotingAccessRepositoryImpl{}
}

type VotingAccessRepositoryImpl struct{}

func (repository *VotingAccessRepositoryImpl) Create(ctx context.Context, tx *sql.Tx, votingAccess domain.VotingAccess) error {
	SQL := `
	INSERT INTO voting_access (user_id, hashed)
	VALUES ($1, $2)
	`

	_, err := tx.ExecContext(ctx, SQL, votingAccess.UserId, votingAccess.Hashed)
	if err != nil {
		return err
	}

	return nil
}

func (repository *VotingAccessRepositoryImpl) Update(ctx context.Context, tx *sql.Tx, votingAccess domain.VotingAccess) error {
	SQL := `
	UPDATE voting_access
	SET hashed = $1
	WHERE user_id = $2
	`

	_, err := tx.ExecContext(ctx, SQL, votingAccess.Hashed, votingAccess.UserId)
	if err != nil {
		return err
	}

	return nil
}

func (repository *VotingAccessRepositoryImpl) GetByUserId(ctx context.Context, tx *sql.Tx, userId int) (domain.VotingAccess, error) {
	SQL := `
	SELECT user_id, hashed
	FROM voting_access
	WHERE user_id = $1
	`

	var votingAccess domain.VotingAccess
	err := tx.QueryRowContext(ctx, SQL, userId).Scan(&votingAccess.UserId, &votingAccess.Hashed)
	if err != nil {
		return domain.VotingAccess{}, err
	}

	return votingAccess, nil
}

func (repository *VotingAccessRepositoryImpl) IsUserEverVoted(ctx context.Context, tx *sql.Tx, userId int) (bool, error) {
	var isVoted bool

	SQL := `
	SELECT EXISTS (
		SELECT 1
		FROM votes
		WHERE hashed_nim = (
			SELECT hashed FROM voting_access WHERE user_id = $1
		)
	)
	`

	err := tx.QueryRowContext(ctx, SQL, userId).Scan(&isVoted)
	if err != nil {
		return isVoted, err
	}

	return isVoted, nil
}

func (repository *VotingAccessRepositoryImpl) DeleteByUserId(ctx context.Context, tx *sql.Tx, userId int) error {
	SQL := `
	DELETE FROM voting_access
	WHERE user_id = $1
	`

	_, err := tx.ExecContext(ctx, SQL, userId)
	if err != nil {
		return err
	}

	return nil
}

func (repository *VotingAccessRepositoryImpl) CreateBulk(ctx context.Context, tx *sql.Tx, votingAccesses []domain.VotingAccess) error {
	var (
		queryBuilder strings.Builder
		args         []interface{}
	)

	queryBuilder.WriteString("INSERT INTO voting_access (user_id, hashed) VALUES ")

	for i, votingAccess := range votingAccesses {
		start := i*2 + 1

		queryBuilder.WriteString(fmt.Sprintf("($%d, $%d)", start, start+1))

		if i < len(votingAccesses)-1 {
			queryBuilder.WriteString(", ")
		}

		args = append(args, votingAccess.UserId, votingAccess.Hashed)
	}

	query := queryBuilder.String()

	_, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
