package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/HIMA-TI-e-Election/model/domain"
)

type UserRepository interface {
	Save(ctx context.Context, tx pgx.Tx, user domain.User) (domain.User, error)
	UpdatePartial(ctx context.Context, tx pgx.Tx, id int, updates map[string]interface{}) (domain.User, error)
	GetByNIM(ctx context.Context, tx pgx.Tx, nim string) (domain.User, error)
	GetUserBySession(ctx context.Context, tx pgx.Tx, sessionId string) (domain.User, error)
	GetAdmins(ctx context.Context, tx pgx.Tx) ([]domain.User, error)
	GetById(ctx context.Context, tx pgx.Tx, userId int) (domain.User, error)
	GetAll(ctx context.Context, tx pgx.Tx) ([]domain.User, error)
	GetByIdWithPassword(ctx context.Context, tx pgx.Tx, userId int) (domain.User, error)
	UpdateById(ctx context.Context, tx pgx.Tx, userId int, user domain.User) (domain.User, error)
	DeleteById(ctx context.Context, tx pgx.Tx, userId int) error
	GetByPhoneNumber(ctx context.Context, tx pgx.Tx, phoneNumber string) (domain.User, error)
	SaveBulk(ctx context.Context, tx pgx.Tx, users []domain.User) ([]domain.User, error)
	GetUsersWithNullPassword(ctx context.Context, tx pgx.Tx) ([]domain.User, error)
	UpdateBulk(ctx context.Context, tx pgx.Tx, users []domain.User) error
}
