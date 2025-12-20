package repository

import (
	"context"
	"database/sql"

	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/domain"
)

func NewAuthRepository() AuthRepository {
	return &AuthRepositoryImpl{}
}

type AuthRepositoryImpl struct{}

func (repository *AuthRepositoryImpl) Create(ctx context.Context, tx *sql.Tx, session domain.Session) (domain.Session, error) {
	SQL := `
	INSERT INTO sessions (session_id, user_id, max_age_seconds)
	VALUES ($1, $2, $3)
	RETURNING created_at
	`

	err := tx.QueryRowContext(ctx, SQL, session.SessionId, session.UserId, session.MaxAgeSeconds).Scan(&session.CreatedAt)
	if err != nil {
		return domain.Session{}, err
	}

	return session, nil
}

func (repository *AuthRepositoryImpl) GetSessionById(ctx context.Context, tx *sql.Tx, sessionId string) (domain.Session, error) {
	SQL := `
	SELECT session_id, user_id, created_at, max_age_seconds
	FROM sessions
	WHERE session_id = $1
	`

	var session domain.Session
	err := tx.QueryRowContext(ctx, SQL, sessionId).Scan(&session.SessionId, &session.UserId, &session.CreatedAt, &session.MaxAgeSeconds)
	if err != nil {
		return domain.Session{}, err
	}

	return session, nil
}

func (repository *AuthRepositoryImpl) Delete(ctx context.Context, tx *sql.Tx, sessionId string) error {
	SQL := `
	DELETE FROM sessions
	WHERE session_id = $1
	`

	_, err := tx.ExecContext(ctx, SQL, sessionId)
	if err != nil {
		return err
	}

	return nil
}
