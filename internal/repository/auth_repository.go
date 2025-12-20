package repository

import (
	"context"
	"database/sql"

	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/domain"
)

type AuthRepository interface {
	Create(ctx context.Context, tx *sql.Tx, session domain.Session) (domain.Session, error)
	GetSessionById(ctx context.Context, tx *sql.Tx, sessionId string) (domain.Session, error)
	Delete(ctx context.Context, tx *sql.Tx, sessionId string) error
}
