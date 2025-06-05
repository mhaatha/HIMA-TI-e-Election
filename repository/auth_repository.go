package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/HIMA-TI-e-Election/model/domain"
)

type AuthRepository interface {
	Create(ctx context.Context, tx pgx.Tx, session domain.Session) (domain.Session, error)
	GetSessionById(ctx context.Context, tx pgx.Tx, sessionId string) (domain.Session, error)
	Delete(ctx context.Context, tx pgx.Tx, sessionId string) error
}
