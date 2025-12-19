package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	appError "github.com/mhaatha/HIMA-TI-e-Election/internal/errors"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/helper"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/domain"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/web"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/repository"
)

func NewAuthService(authRepository repository.AuthRepository, userService UserService, db *sql.DB, validate *validator.Validate) AuthService {
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
	DB             *sql.DB
	Validate       *validator.Validate
}

func (service *AuthServiceImpl) LoginUser(ctx context.Context, maxAge int, request web.LoginRequest) (web.LoginResponse, string, error) {
	// Validate request
	err := service.Validate.Struct(request)
	if err != nil {
		return web.LoginResponse{}, "", appError.NewAppError(
			http.StatusBadRequest,
			"Invalid request payload",
			appError.FormatValidationDetails(err.(validator.ValidationErrors)),
			fmt.Errorf("%w: %v", appError.ErrValidation, err),
		)
	}

	// Open Transaction
	tx, err := service.DB.BeginTx(ctx, nil)
	if err != nil {
		return web.LoginResponse{}, "", appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.RollbackQuietly(tx)

	// Get user by NIM to check password
	user, err := service.UserService.GetByNIM(ctx, request.NIM)
	if err != nil {
		return web.LoginResponse{}, "", appError.NewAppError(
			http.StatusUnauthorized,
			"Invalid credentials",
			"The NIM or password is incorrect",
			err,
		)
	}

	// Validate NIM and password
	if !helper.CheckPasswordHash(user.Password, request.Password) {
		return web.LoginResponse{}, "", appError.NewAppError(
			http.StatusUnauthorized,
			"Invalid credentials",
			"The NIM or password is incorrect",
			fmt.Errorf("%w", appError.ErrInvalidCredentials),
		)
	}

	// Save to sessions db
	session, err := service.AuthRepository.Create(ctx, tx, domain.Session{
		SessionId:     helper.Base64SessionId(),
		UserId:        user.ID,
		MaxAgeSeconds: maxAge,
	})
	if err != nil {
		return web.LoginResponse{}, "", appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to create session: %w", err),
		)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return web.LoginResponse{}, "", appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("transaction commit failed: %w", err),
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
		return web.LoginResponse{}, "", appError.NewAppError(
			http.StatusBadRequest,
			"Invalid request payload",
			appError.FormatValidationDetails(err.(validator.ValidationErrors)),
			fmt.Errorf("%w: %v", appError.ErrValidation, err),
		)
	}

	// Open Transaction
	tx, err := service.DB.BeginTx(ctx, nil)
	if err != nil {
		return web.LoginResponse{}, "", appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.RollbackQuietly(tx)

	// Check if any password matches
	admins, err := service.UserService.GetAdmins(ctx)
	if err != nil {
		return web.LoginResponse{}, "", appError.NewAppError(
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
				return web.LoginResponse{}, "", appError.NewAppError(
					http.StatusInternalServerError,
					"Internal Server Error",
					"Failed to process your request due to an unexpected error. Please try again later.",
					fmt.Errorf("failed to create session: %w", err),
				)
			}

			// Commit transaction
			if err = tx.Commit(); err != nil {
				return web.LoginResponse{}, "", appError.NewAppError(
					http.StatusInternalServerError,
					"Internal Server Error",
					"Failed to process your request due to an unexpected error. Please try again later.",
					fmt.Errorf("transaction commit failed: %w", err),
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

	return web.LoginResponse{}, "", appError.NewAppError(
		http.StatusUnauthorized,
		"Invalid credentials",
		"Invalid NIM or password",
		fmt.Errorf("invalid credensials: %w", appError.ErrInvalidCredentials),
	)
}

func (service *AuthServiceImpl) Logout(ctx context.Context, sessionId string) error {
	// Open Transaction
	tx, err := service.DB.BeginTx(ctx, nil)
	if err != nil {
		return appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.RollbackQuietly(tx)

	// Delete session
	err = service.AuthRepository.Delete(ctx, tx, sessionId)
	if err != nil {
		return appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to delete session: %w", err),
		)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("transaction commit failed: %w", err),
		)
	}

	return nil
}

func (service *AuthServiceImpl) UserValidateSession(ctx context.Context, sessionId string) (web.SessionResponse, error) {
	// Open Transaction
	tx, err := service.DB.BeginTx(ctx, nil)
	if err != nil {
		return web.SessionResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.RollbackQuietly(tx)

	// Get session
	session, err := service.AuthRepository.GetSessionById(ctx, tx, sessionId)
	if err != nil {
		return web.SessionResponse{}, appError.NewAppError(
			http.StatusUnauthorized,
			"Invalid session data",
			"Session data may be corrupted, missing or expired",
			fmt.Errorf("%w: session with id '%v' is not found", appError.ErrSessionNotFound, sessionId),
		)
	}

	// Validate session
	if session.CreatedAt.Add(time.Second * time.Duration(session.MaxAgeSeconds)).Before(time.Now()) {
		return web.SessionResponse{}, appError.NewAppError(
			http.StatusUnauthorized,
			"Invalid session data",
			"Session data may be corrupted, missing or expired",
			fmt.Errorf("%w: session with id '%v' is expired", appError.ErrSessionExpired, sessionId),
		)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return web.SessionResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("transaction commit failed: %w", err),
		)
	}

	return helper.ToSessionResponse(session), nil
}

func (service *AuthServiceImpl) AdminValidateSession(ctx context.Context, sessionId string) (web.SessionResponse, error) {
	// Open Transaction
	tx, err := service.DB.BeginTx(ctx, nil)
	if err != nil {
		return web.SessionResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.RollbackQuietly(tx)

	// Get session
	session, err := service.AuthRepository.GetSessionById(ctx, tx, sessionId)
	if err != nil {
		return web.SessionResponse{}, appError.NewAppError(
			http.StatusUnauthorized,
			"Invalid session data",
			"Session data may be corrupted, missing or expired",
			fmt.Errorf("%w: session with id '%v' is not found", appError.ErrSessionNotFound, sessionId),
		)
	}

	// Check if user is admin
	user, _ := service.UserService.GetCurrent(ctx, sessionId)

	if user.Role != "admin" {
		return web.SessionResponse{}, appError.NewAppError(
			http.StatusForbidden,
			"Forbidden access",
			"You do not have permission to access this resource",
			fmt.Errorf("%w: user with id '%v' is not an admin", appError.ErrForbiddenAccess, user.ID),
		)
	}

	// Validate session
	if session.CreatedAt.Add(time.Second * time.Duration(session.MaxAgeSeconds)).Before(time.Now()) {
		return web.SessionResponse{}, appError.NewAppError(
			http.StatusUnauthorized,
			"Invalid session data",
			"Session data may be corrupted, missing or expired",
			fmt.Errorf("%w: session with id '%v' is expired", appError.ErrSessionExpired, sessionId),
		)
	}

	// Commit transcation
	if err = tx.Commit(); err != nil {
		return web.SessionResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to commit transaction",
			fmt.Errorf("transaction commit failed: %w", err),
		)
	}

	return helper.ToSessionResponse(session), nil
}
