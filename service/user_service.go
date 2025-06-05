package service

import (
	"context"

	"github.com/mhaatha/HIMA-TI-e-Election/model/web"
)

type UserService interface {
	Create(ctx context.Context, request web.UserCreateRequest) (web.UserResponse, error)
	UpdateCurrent(ctx context.Context, sessionId string, request web.UserUpdateCurrentRequest) (web.UserResponse, error)
	GetCurrent(ctx context.Context, sessionId string) (web.UserResponse, error)
	GetByNIM(ctx context.Context, nim string) (web.UserGetByNimResponse, error)
	GetAdmins(ctx context.Context) ([]web.AdminResponse, error)
	GetById(ctx context.Context, userId int) (web.UserResponse, error)
	GetAll(ctx context.Context) ([]web.UserResponse, error)
	UpdateById(ctx context.Context, userId int, request web.UserUpdateByIdRequest) (web.UserResponse, error)
	DeleteById(ctx context.Context, userId int) error
	CreateBulk(ctx context.Context, request []web.UserCreateRequest) ([]web.UserResponse, error)
	GeneratePassword(ctx context.Context) error
}
