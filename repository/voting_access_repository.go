package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/HIMA-TI-e-Election/model/domain"
)

type VotingAccessRepository interface {
	Create(ctx context.Context, tx pgx.Tx, votingAccess domain.VotingAccess) error
	Update(ctx context.Context, tx pgx.Tx, votingAccess domain.VotingAccess) error
	GetByUserId(ctx context.Context, tx pgx.Tx, userId int) (domain.VotingAccess, error)
	IsUserEverVoted(ctx context.Context, tx pgx.Tx, userId int) (bool, error)
	DeleteByUserId(ctx context.Context, tx pgx.Tx, userId int) error
	CreateBulk(ctx context.Context, tx pgx.Tx, votingAccesses []domain.VotingAccess) error
}
