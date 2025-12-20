package repository

import (
	"context"
	"database/sql"

	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/domain"
)

type UserRepository interface {
	Save(ctx context.Context, tx *sql.Tx, user domain.User) (domain.User, error)
	UpdatePartial(ctx context.Context, tx *sql.Tx, id int, updates map[string]interface{}) (domain.User, error)
	GetByNIM(ctx context.Context, tx *sql.Tx, nim string) (domain.User, error)
	GetUserBySession(ctx context.Context, tx *sql.Tx, sessionId string) (domain.User, error)
	GetAdmins(ctx context.Context, tx *sql.Tx) ([]domain.User, error)
	GetById(ctx context.Context, tx *sql.Tx, userId int) (domain.User, error)
	GetAll(ctx context.Context, tx *sql.Tx) ([]domain.User, error)
	GetByIdWithPassword(ctx context.Context, tx *sql.Tx, userId int) (domain.User, error)
	UpdateById(ctx context.Context, tx *sql.Tx, userId int, user domain.User) (domain.User, error)
	DeleteById(ctx context.Context, tx *sql.Tx, userId int) error
	GetByPhoneNumber(ctx context.Context, tx *sql.Tx, phoneNumber string) (domain.User, error)
	SaveBulk(ctx context.Context, tx *sql.Tx, users []domain.User) ([]domain.User, error)
	GetUsersWithNullPassword(ctx context.Context, tx *sql.Tx) ([]domain.User, error)
	UpdateBulk(ctx context.Context, tx *sql.Tx, users []domain.User) error
}
