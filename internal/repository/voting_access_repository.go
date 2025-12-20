package repository

import (
	"context"
	"database/sql"

	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/domain"
)

type VotingAccessRepository interface {
	Create(ctx context.Context, tx *sql.Tx, votingAccess domain.VotingAccess) error
	Update(ctx context.Context, tx *sql.Tx, votingAccess domain.VotingAccess) error
	GetByUserId(ctx context.Context, tx *sql.Tx, userId int) (domain.VotingAccess, error)
	IsUserEverVoted(ctx context.Context, tx *sql.Tx, userId int) (bool, error)
	DeleteByUserId(ctx context.Context, tx *sql.Tx, userId int) error
	CreateBulk(ctx context.Context, tx *sql.Tx, votingAccesses []domain.VotingAccess) error
}
