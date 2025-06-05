package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mhaatha/HIMA-TI-e-Election/errors"
	"github.com/mhaatha/HIMA-TI-e-Election/helper"
	"github.com/mhaatha/HIMA-TI-e-Election/model/domain"
	"github.com/mhaatha/HIMA-TI-e-Election/model/web"
	"github.com/mhaatha/HIMA-TI-e-Election/repository"
)

func NewAuthService(authRepository repository.AuthRepository, userService UserService, db *pgxpool.Pool, validate *validator.Validate) AuthService {
	return &AuthServiceImpl{
		AuthRepository: authRepository,
		UserService:    userService,
		DB:             db,
		Validate:       validate,
	}
}

type AuthServiceImpl struct {
	AuthRepository repository.AuthRepository
	UserService    UserService
	DB             *pgxpool.Pool
	Validate       *validator.Validate
}

func (service *AuthServiceImpl) LoginUser(ctx context.Context, maxAge int, request web.LoginRequest) (web.LoginResponse, string, error) {
	// Validate request
	err := service.Validate.Struct(request)
	if err != nil {
		return web.LoginResponse{}, "", errors.NewAppError(
			http.StatusBadRequest,
			"Invalid request payload",
			errors.FormatValidationDetails(err.(validator.ValidationErrors)),
			fmt.Errorf("%w: %v", errors.ErrValidation, err),
		)
	}

	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return web.LoginResponse{}, "", errors.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", errors.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Get user by NIM to check password
	user, err := service.UserService.GetByNIM(ctx, request.NIM)
	if err != nil {
		return web.LoginResponse{}, "", errors.NewAppError(
			http.StatusUnauthorized,
			"Invalid credentials",
			"The NIM or password is incorrect",
			err,
		)
	}

	// Validate NIM and password
	if !helper.CheckPasswordHash(user.Password, request.Password) {
		return web.LoginResponse{}, "", errors.NewAppError(
			http.StatusUnauthorized,
			"Invalid credentials",
			"The NIM or password is incorrect",
			fmt.Errorf("%w", errors.ErrInvalidCredentials),
		)
	}

	// Save to sessions db
	session, err := service.AuthRepository.Create(ctx, tx, domain.Session{
		SessionId:     helper.Base64SessionId(),
		UserId:        user.ID,
		MaxAgeSeconds: maxAge,
	})
	if err != nil {
		return web.LoginResponse{}, "", errors.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to create session: %w", err),
		)
	}

	// Write response
	serviceResponse := domain.User{
		Id:           user.ID,
		NIM:          user.NIM,
		FullName:     user.FullName,
		StudyProgram: user.StudyProgram,
		Role:         user.Role,
		PhoneNumber:  user.PhoneNumber,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	return helper.ToLoginResponse(serviceResponse), session.SessionId, nil
}

func (service *AuthServiceImpl) LoginAdmin(ctx context.Context, maxAge int, request web.LoginRequest) (web.LoginResponse, string, error) {
	// Validate request
	err := service.Validate.Struct(request)
	if err != nil {
		return web.LoginResponse{}, "", errors.NewAppError(
			http.StatusBadRequest,
			"Invalid request payload",
			errors.FormatValidationDetails(err.(validator.ValidationErrors)),
			fmt.Errorf("%w: %v", errors.ErrValidation, err),
		)
	}

	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return web.LoginResponse{}, "", errors.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", errors.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Check if any password matches
	admins, err := service.UserService.GetAdmins(ctx)
	if err != nil {
		return web.LoginResponse{}, "", errors.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to login: %w", err),
		)
	}

	for _, admin := range admins {
		if helper.CheckPasswordHash(admin.Password, request.Password) {
			// Save to sessions db
			session, err := service.AuthRepository.Create(ctx, tx, domain.Session{
				SessionId:     helper.Base64SessionId(),
				UserId:        admin.ID,
				MaxAgeSeconds: maxAge,
			})
			if err != nil {
				return web.LoginResponse{}, "", errors.NewAppError(
					http.StatusInternalServerError,
					"Internal Server Error",
					"Failed to process your request due to an unexpected error. Please try again later.",
					fmt.Errorf("failed to create session: %w", err),
				)
			}

			// Write response
			serviceResponse := domain.User{
				Id:           admin.ID,
				NIM:          admin.NIM,
				FullName:     admin.FullName,
				StudyProgram: admin.StudyProgram,
				Role:         admin.Role,
				PhoneNumber:  admin.PhoneNumber,
				CreatedAt:    admin.CreatedAt,
				UpdatedAt:    admin.UpdatedAt,
			}

			return helper.ToLoginResponse(serviceResponse), session.SessionId, nil
		}
	}

	return web.LoginResponse{}, "", errors.NewAppError(
		http.StatusUnauthorized,
		"Invalid credentials",
		"Invalid NIM or password",
		fmt.Errorf("%w", errors.ErrInvalidCredentials),
	)
}

func (service *AuthServiceImpl) Logout(ctx context.Context, sessionId string) error {
	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return errors.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", errors.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Delete session
	err = service.AuthRepository.Delete(ctx, tx, sessionId)
	if err != nil {
		return errors.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to delete session: %w", err),
		)
	}

	return nil
}

func (service *AuthServiceImpl) UserValidateSession(ctx context.Context, sessionId string) (web.SessionResponse, error) {
	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return web.SessionResponse{}, errors.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", errors.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Get session
	session, err := service.AuthRepository.GetSessionById(ctx, tx, sessionId)
	if err != nil {
		return web.SessionResponse{}, errors.NewAppError(
			http.StatusUnauthorized,
			"Invalid session data",
			"Session data may be corrupted, missing or expired",
			fmt.Errorf("%w: session with id '%v' is not found", errors.ErrSessionNotFound, sessionId),
		)
	}

	// Validate session
	if session.CreatedAt.Add(time.Second * time.Duration(session.MaxAgeSeconds)).Before(time.Now()) {
		return web.SessionResponse{}, errors.NewAppError(
			http.StatusUnauthorized,
			"Invalid session data",
			"Session data may be corrupted, missing or expired",
			fmt.Errorf("%w: session with id '%v' is expired", errors.ErrSessionExpired, sessionId),
		)
	}

	return helper.ToSessionResponse(session), nil
}

func (service *AuthServiceImpl) AdminValidateSession(ctx context.Context, sessionId string) (web.SessionResponse, error) {
	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return web.SessionResponse{}, errors.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", errors.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Get session
	session, err := service.AuthRepository.GetSessionById(ctx, tx, sessionId)
	if err != nil {
		return web.SessionResponse{}, errors.NewAppError(
			http.StatusUnauthorized,
			"Invalid session data",
			"Session data may be corrupted, missing or expired",
			fmt.Errorf("%w: session with id '%v' is not found", errors.ErrSessionNotFound, sessionId),
		)
	}

	// Check if user is admin
	user, _ := service.UserService.GetCurrent(ctx, sessionId)

	if user.Role != "admin" {
		return web.SessionResponse{}, errors.NewAppError(
			http.StatusForbidden,
			"Forbidden access",
			"You do not have permission to access this resource",
			fmt.Errorf("%w: user with id '%v' is not an admin", errors.ErrForbiddenAccess, user.ID),
		)
	}

	// Validate session
	if session.CreatedAt.Add(time.Second * time.Duration(session.MaxAgeSeconds)).Before(time.Now()) {
		return web.SessionResponse{}, errors.NewAppError(
			http.StatusUnauthorized,
			"Invalid session data",
			"Session data may be corrupted, missing or expired",
			fmt.Errorf("%w: session with id '%v' is expired", errors.ErrSessionExpired, sessionId),
		)
	}

	return helper.ToSessionResponse(session), nil
}
