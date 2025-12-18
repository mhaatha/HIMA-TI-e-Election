package service

import (
	"context"

	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/web"
)

type AuthService interface {
	LoginUser(ctx context.Context, maxAge int, request web.LoginRequest) (web.LoginResponse, string, error)
	LoginAdmin(ctx context.Context, maxAge int, request web.LoginRequest) (web.LoginResponse, string, error)
	Logout(ctx context.Context, sessionId string) error
	UserValidateSession(ctx context.Context, sessionId string) (web.SessionResponse, error)
	AdminValidateSession(ctx context.Context, sessionId string) (web.SessionResponse, error)
}
