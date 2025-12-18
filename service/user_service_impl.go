package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	envConfig "github.com/mhaatha/HIMA-TI-e-Election/config"
	appError "github.com/mhaatha/HIMA-TI-e-Election/errors"
	"github.com/mhaatha/HIMA-TI-e-Election/helper"
	"github.com/mhaatha/HIMA-TI-e-Election/model/domain"
	"github.com/mhaatha/HIMA-TI-e-Election/model/web"
	"github.com/mhaatha/HIMA-TI-e-Election/repository"
)

func NewUserService(userRepository repository.UserRepository, votingAccessRepository repository.VotingAccessRepository, envConfig *envConfig.Config, db *pgxpool.Pool, validate *validator.Validate) UserService {
	return &UserServiceImpl{
		UserRepository:         userRepository,
		VotingAccessRepository: votingAccessRepository,
		EnvConfig:              envConfig,
		DB:                     db,
		Validate:               validate,
	}
}

type UserServiceImpl struct {
	UserRepository         repository.UserRepository
	VotingAccessRepository repository.VotingAccessRepository
	EnvConfig              *envConfig.Config
	DB                     *pgxpool.Pool
	Validate               *validator.Validate
}

func (service *UserServiceImpl) Create(ctx context.Context, request web.UserCreateRequest) (web.UserResponse, error) {
	// Validate request
	err := service.Validate.Struct(request)
	if err != nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusBadRequest,
			"Invalid request payload",
			appError.FormatValidationDetails(err.(validator.ValidationErrors)),
			fmt.Errorf("%w: %v", appError.ErrValidation, err),
		)
	}

	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Check if NIM already exists
	user, err := service.UserRepository.GetByNIM(ctx, tx, request.NIM)
	if err == nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusBadRequest,
			"NIM already exists",
			"The NIM you provided already exists. Try another NIM or try logging in instead.",
			fmt.Errorf("user with NIM %s already exists", request.NIM),
		)
	}

	// Check if phone_number already exists
	user, err = service.UserRepository.GetByPhoneNumber(ctx, tx, request.PhoneNumber)
	if err == nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusBadRequest,
			"Phone number already exists",
			"The phone number you provided already exists. Try another phone number.",
			fmt.Errorf("user with phone number %s already exists", request.PhoneNumber),
		)
	}

	// Hash the password before saving it to database
	hashedPassword, err := helper.HashPassword(request.Password)
	if err != nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to hash password: %w", err),
		)
	}

	user = domain.User{
		NIM:          request.NIM,
		FullName:     request.FullName,
		StudyProgram: request.StudyProgram,
		Password:     hashedPassword,
		Role:         request.Role,
		PhoneNumber:  request.PhoneNumber,
	}

	// Save to database
	user, err = service.UserRepository.Save(ctx, tx, user)
	if err != nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to create user: %w", err),
		)
	}

	// Only role with student can have voting access
	if user.Role == "student" {
		err := service.VotingAccessRepository.Create(ctx, tx, domain.VotingAccess{
			UserId: user.Id,
			Hashed: helper.HashNIM(user.NIM),
		})

		if err != nil {
			return web.UserResponse{}, appError.NewAppError(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Failed to process your request due to an unexpected error. Please try again later.",
				fmt.Errorf("failed to create voting_access: %w", err),
			)
		}
	}

	return helper.ToUserResponse(user), nil
}

func (service *UserServiceImpl) UpdateCurrent(ctx context.Context, sessionId string, request web.UserUpdateCurrentRequest) (web.UserResponse, error) {
	// Validate request
	err := service.Validate.Struct(request)
	if err != nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusBadRequest,
			"Invalid request payload",
			appError.FormatValidationDetails(err.(validator.ValidationErrors)),
			fmt.Errorf("%w: %v", appError.ErrValidation, err),
		)
	}

	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Create map of fields to update
	updates := make(map[string]interface{})

	// Add fields to update map if they are provided in the request
	if request.NIM != "" {
		updates["nim"] = request.NIM
	}
	if request.FullName != "" {
		updates["full_name"] = request.FullName
	}
	if request.StudyProgram != "" {
		updates["study_program"] = request.StudyProgram
	}

	// Call repository to get current user
	user, err := service.UserRepository.GetUserBySession(ctx, tx, sessionId)
	if err != nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusNotFound,
			"User not found",
			"The user you are trying to update does not exist.",
			fmt.Errorf("failed to get user by session: %w", err),
		)
	}

	// If NIM is provided, check if it already used by another user
	if user.NIM != request.NIM {
		_, err := service.UserRepository.GetByNIM(ctx, tx, request.NIM)
		// if err == nil it means NIM already used by another user
		if err == nil {
			return web.UserResponse{}, appError.NewAppError(
				http.StatusBadRequest,
				"NIM already exists",
				"The NIM you provided already exists. Try another NIM or try logging in instead.",
				fmt.Errorf("user with NIM %s already exists: %w", request.NIM, appError.ErrNIMAlreadyExists),
			)
		}
	}

	// Call repository to update current user
	updatedUser, err := service.UserRepository.UpdatePartial(ctx, tx, user.Id, updates)
	if err != nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusNotFound,
			"User not found",
			"The user you are trying to update does not exist.",
			fmt.Errorf("failed to update user: %w", err),
		)
	}

	// Update hashed key in voting_access table
	if user.NIM != request.NIM {
		err := service.VotingAccessRepository.Update(ctx, tx, domain.VotingAccess{
			UserId: user.Id,
			Hashed: helper.HashNIM(request.NIM),
		})

		if err != nil {
			return web.UserResponse{}, appError.NewAppError(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Failed to process your request due to an unexpected error. Please try again later.",
				fmt.Errorf("failed to create voting_access: %w", err),
			)
		}
	}

	return helper.ToUserResponse(updatedUser), nil
}

func (service *UserServiceImpl) GetCurrent(ctx context.Context, sessionId string) (web.UserResponse, error) {
	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Check if user exists
	user, err := service.UserRepository.GetUserBySession(ctx, tx, sessionId)
	if err != nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusNotFound,
			"User not found",
			"The user you are trying to get does not exist.",
			fmt.Errorf("failed to get user by session: %w", err),
		)
	}

	return helper.ToUserResponse(user), nil
}

func (service *UserServiceImpl) GetByNIM(ctx context.Context, nim string) (web.UserGetByNimResponse, error) {
	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return web.UserGetByNimResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Check if NIM exists
	user, err := service.UserRepository.GetByNIM(ctx, tx, nim)
	if err != nil {
		return web.UserGetByNimResponse{}, appError.NewAppError(
			http.StatusNotFound,
			"User not found",
			"The user you are trying to get does not exist.",
			fmt.Errorf("%w: user with NIM %s is not found", appError.ErrNIMNotFound, nim),
		)
	}

	return helper.ToUserGetByNimResponse(user), nil
}

func (service *UserServiceImpl) GetAdmins(ctx context.Context) ([]web.AdminResponse, error) {
	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return []web.AdminResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Get admins
	admins, err := service.UserRepository.GetAdmins(ctx, tx)
	if err != nil {
		return []web.AdminResponse{}, appError.NewAppError(
			http.StatusNotFound,
			"Admins not found",
			"The admins you are trying to get do not exist.",
			fmt.Errorf("failed to get admins: %w", err),
		)
	}

	return helper.ToAdminResponses(admins), nil
}

func (service *UserServiceImpl) GetById(ctx context.Context, userId int) (web.UserResponse, error) {
	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Check if user exists
	user, err := service.UserRepository.GetById(ctx, tx, userId)
	if err != nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusNotFound,
			"User not found",
			"The user you are trying to get does not exist.",
			fmt.Errorf("failed to get user by id: %w", err),
		)
	}

	return helper.ToUserResponse(user), nil
}

func (service *UserServiceImpl) GetAll(ctx context.Context) ([]web.UserResponse, error) {
	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return []web.UserResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Get users
	users, err := service.UserRepository.GetAll(ctx, tx)
	if err != nil {
		return []web.UserResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}

	return helper.ToUserResponses(users), nil
}

func (service *UserServiceImpl) UpdateById(ctx context.Context, userId int, request web.UserUpdateByIdRequest) (web.UserResponse, error) {
	// Validate request
	err := service.Validate.Struct(request)
	if err != nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusBadRequest,
			"Invalid request payload",
			appError.FormatValidationDetails(err.(validator.ValidationErrors)),
			fmt.Errorf("%w: %v", appError.ErrValidation, err),
		)
	}

	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Check if user exists
	user, err := service.UserRepository.GetByIdWithPassword(ctx, tx, userId)
	if err != nil {
		// If user not found
		if errors.Is(err, pgx.ErrNoRows) {
			return web.UserResponse{}, appError.NewAppError(
				http.StatusNotFound,
				"User not found",
				fmt.Sprintf("User with id %v does not exist", userId),
				fmt.Errorf("user with id %v not found: %v", userId, err),
			)
		}

		return web.UserResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to get user with id %v: %v", userId, err),
		)
	}

	// Check is user ever voted
	isVoted, err := service.VotingAccessRepository.IsUserEverVoted(ctx, tx, userId)
	if err != nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to check if user has voted: %v", err),
		)
	}
	if isVoted {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusConflict,
			"User has already voted",
			"Can't update user because user has already voted",
			fmt.Errorf("user has already voted in: %v", err),
		)
	}

	// If NIM is provided, check if it already used by another user
	if user.NIM != request.NIM {
		_, err := service.UserRepository.GetByNIM(ctx, tx, request.NIM)
		// if err == nil it means NIM already used by another user
		if err == nil {
			return web.UserResponse{}, appError.NewAppError(
				http.StatusBadRequest,
				"NIM already exists",
				"The NIM you provided already exists. Try another NIM or try logging in instead.",
				fmt.Errorf("user with NIM %s already exists: %w", request.NIM, appError.ErrNIMAlreadyExists),
			)
		}
	}

	// If phone_number is provided, check if i already used by another user
	if user.PhoneNumber != request.PhoneNumber {
		_, err := service.UserRepository.GetByPhoneNumber(ctx, tx, request.PhoneNumber)
		// if err == nil it means phone number already used by another user
		if err == nil {
			return web.UserResponse{}, appError.NewAppError(
				http.StatusBadRequest,
				"Phone number already exists",
				"The phone number you provided already exists. Try another phone number.",
				fmt.Errorf("user with phone number %s already exists", request.PhoneNumber),
			)
		}
	}

	// Delete the existed voting_access if the old user is a student
	if user.Role == "student" {
		err = service.VotingAccessRepository.DeleteByUserId(ctx, tx, userId)
		if err != nil {
			return web.UserResponse{}, appError.NewAppError(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Failed to process your request due to an unexpected error. Please try again later.",
				fmt.Errorf("failed to delete voting_access: %v", err),
			)
		}
	}

	// Check the request body, if not empty, swap the user to the request body
	if request.NIM != "" {
		user.NIM = request.NIM
	}
	if request.FullName != "" {
		user.FullName = request.FullName
	}
	if request.StudyProgram != "" {
		user.StudyProgram = request.StudyProgram
	}
	if request.Password != "" {
		// Hash the password
		hashedPassword, err := helper.HashPassword(request.Password)
		if err != nil {
			return web.UserResponse{}, appError.NewAppError(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Failed to process your request due to an unexpected error. Please try again later.",
				fmt.Errorf("failed to hash password: %w", err),
			)
		}

		user.Password = hashedPassword
	}
	if request.Role != "" {
		user.Role = request.Role
	}
	if request.PhoneNumber != "" {
		user.PhoneNumber = request.PhoneNumber
	}

	// Create a new voting_access if the new user is a student
	// Only role with student can have voting access
	if user.Role == "student" {
		err := service.VotingAccessRepository.Create(ctx, tx, domain.VotingAccess{
			UserId: user.Id,
			Hashed: helper.HashNIM(user.NIM),
		})

		if err != nil {
			return web.UserResponse{}, appError.NewAppError(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Failed to process your request due to an unexpected error. Please try again later.",
				fmt.Errorf("failed to create voting_access: %w", err),
			)
		}
	}

	// Call repository
	userResponse, err := service.UserRepository.UpdateById(ctx, tx, userId, user)
	if err != nil {
		return web.UserResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to update user with id %v: %v", userId, err),
		)
	}

	return helper.ToUserResponse(userResponse), nil
}

func (service *UserServiceImpl) DeleteById(ctx context.Context, userId int) error {
	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Check if user exists
	_, err = service.UserRepository.GetById(ctx, tx, userId)
	if err != nil {
		// If user not found
		if errors.Is(err, pgx.ErrNoRows) {
			return appError.NewAppError(
				http.StatusNotFound,
				"User not found",
				fmt.Sprintf("User with id %v does not exist", userId),
				fmt.Errorf("user with id %v not found: %v", userId, err),
			)
		}

		return appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to get user with id %v: %v", userId, err),
		)
	}

	// Check is user ever voted
	isVoted, err := service.VotingAccessRepository.IsUserEverVoted(ctx, tx, userId)
	if err != nil {
		return appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to check if user has voted: %v", err),
		)
	}
	if isVoted {
		return appError.NewAppError(
			http.StatusConflict,
			"User has already voted",
			"Can't delete user because user has already voted",
			fmt.Errorf("user has already voted in: %v", err),
		)
	}

	// Call repository
	err = service.UserRepository.DeleteById(ctx, tx, userId)
	if err != nil {
		return appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to update user with id %v: %v", userId, err),
		)
	}

	return nil
}

func (service *UserServiceImpl) CreateBulk(ctx context.Context, request []web.UserCreateRequest) ([]web.UserResponse, error) {
	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return []web.UserResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	var users []domain.User

	// Validate each request
	for i, currentUser := range request {
		// Validate request
		err := service.Validate.Struct(currentUser)
		if err != nil {
			return []web.UserResponse{}, appError.NewAppError(
				http.StatusBadRequest,
				"Invalid request payload",
				appError.FormatValidationDetailsWithRow((err.(validator.ValidationErrors)), i+1),
				fmt.Errorf("%w: %v in csv line %d", appError.ErrValidation, err, i+1),
			)
		}

		// Check if NIM already exists
		_, err = service.UserRepository.GetByNIM(ctx, tx, currentUser.NIM)
		if err == nil {
			return []web.UserResponse{}, appError.NewAppError(
				http.StatusBadRequest,
				"NIM already exists",
				fmt.Sprintf("CSV's line %d: NIM '%s' is already exists", i+1, currentUser.NIM),
				fmt.Errorf("user with NIM %s already exists", currentUser.NIM),
			)
		}

		// Check if phone_number already exists
		_, err = service.UserRepository.GetByPhoneNumber(ctx, tx, currentUser.PhoneNumber)
		if err == nil {
			return []web.UserResponse{}, appError.NewAppError(
				http.StatusBadRequest,
				"Phone number already exists",
				fmt.Sprintf("CSV's line %d: Phone number '%s' is already exists", i+1, currentUser.PhoneNumber),
				fmt.Errorf("user with phone number %s already exists", currentUser.PhoneNumber),
			)
		}

		users = append(users, domain.User{
			NIM:          currentUser.NIM,
			FullName:     currentUser.FullName,
			StudyProgram: currentUser.StudyProgram,
			PhoneNumber:  currentUser.PhoneNumber,
		})
	}

	// Insert user data to database in bulk
	users, err = service.UserRepository.SaveBulk(ctx, tx, users)
	if err != nil {
		return []web.UserResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to create user in bulk: %w", err),
		)
	}

	var votingAccesses []domain.VotingAccess

	// Initialize voting_access
	for _, user := range users {
		if user.Role == "student" {
			votingAccesses = append(votingAccesses, domain.VotingAccess{
				UserId: user.Id,
				Hashed: helper.HashNIM(user.NIM),
			})
		}
	}

	// Insert voting_access data to database in bulk
	err = service.VotingAccessRepository.CreateBulk(ctx, tx, votingAccesses)
	if err != nil {
		return []web.UserResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to create voting_access in bulk: %w", err),
		)
	}

	return helper.ToUserResponses(users), nil
}

func (service *UserServiceImpl) GeneratePassword(ctx context.Context) error {
	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Get all users with null password
	users, err := service.UserRepository.GetUsersWithNullPassword(ctx, tx)
	if err != nil {
		return appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to get users with null password: %w", err),
		)
	}

	if len(users) == 0 {
		envConfig.Log.Info("No users with null password, skipping password generation")
		return appError.NewAppError(
			http.StatusOK,
			"No users with null password",
			"Password generation skipped",
			nil,
		)
	}

	var listOfUsers []domain.User
	/*
		Required data to perform send message via WhatsApp:
		1. FullName
		2. Password
		3. PhoneNumber
	*/
	var willBeSendToWhatsApp []domain.User

	// Generate password for each user
	for _, user := range users {
		// Generate password
		password, err := helper.GeneratePassword(helper.DefaultPasswordLength)
		if err != nil {
			return appError.NewAppError(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Failed to process your request due to an unexpected error. Please try again later.",
				fmt.Errorf("failed to generate password: %w", err),
			)
		}

		// Data that will be send to WA
		willBeSendToWhatsApp = append(willBeSendToWhatsApp, domain.User{
			FullName:    user.FullName,
			Password:    password,
			PhoneNumber: user.PhoneNumber,
		})

		// Hash the password
		hashedPassword, err := helper.HashPassword(password)
		if err != nil {
			return appError.NewAppError(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Failed to process your request due to an unexpected error. Please try again later.",
				fmt.Errorf("failed to hash password: %w", err),
			)
		}

		user.Password = hashedPassword

		listOfUsers = append(listOfUsers, user)
	}

	// Bulk update for all users
	err = service.UserRepository.UpdateBulk(ctx, tx, listOfUsers)
	if err != nil {
		return appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to update user in bulk: %w", err),
		)
	}

	// Send the password to WhatsApp using Fonnte
	URL := "https://api.fonnte.com/send"
	for _, user := range willBeSendToWhatsApp {
		payload := map[string]string{
			"target":  user.PhoneNumber,
			"message": fmt.Sprintf("Halo *%s*, ini adalah password Voting Anda: *%s*\nGunakan password ini untuk login dan berikan suara terbaikmu!\nTerima kasih atas partisipasinya.\n\nVote di sini: https://himati-e-election-polnes.vercel.app", user.FullName, user.Password),
		}

		// Convert the user data to JSON
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return appError.NewAppError(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Failed to process your request due to an unexpected error. Please try again later.",
				fmt.Errorf("failed to convert user data to JSON: %w", err),
			)
		}

		// Create the request
		req, err := http.NewRequest(http.MethodPost, URL, bytes.NewBuffer(jsonPayload))
		if err != nil {
			return appError.NewAppError(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Failed to process your request due to an unexpected error. Please try again later.",
				fmt.Errorf("failed to create request: %w", err),
			)
		}

		// Set the request header
		req.Header.Set("Authorization", service.EnvConfig.FonnteAPIKey)
		req.Header.Set("Content-Type", "application/json")

		// Send the request to Fonnte
		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return appError.NewAppError(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Failed to process your request due to an unexpected error. Please try again later.",
				fmt.Errorf("failed to send request to Fonnte: %w", err),
			)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return appError.NewAppError(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Failed to process your request due to an unexpected error. Please try again later.",
				fmt.Errorf("failed to read response body: %w", err),
			)
		}

		var fonnteResp map[string]interface{}
		if err := json.Unmarshal(body, &fonnteResp); err != nil {
			return appError.NewAppError(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Failed to process your request due to an unexpected error. Please try again later.",
				fmt.Errorf("failed to unmarshal response body: %w", err),
			)
		}

		if !fonnteResp["status"].(bool) {
			// Rollback the transaction
			if err := tx.Rollback(ctx); err != nil {
				return appError.NewAppError(
					http.StatusInternalServerError,
					"Internal Server Error",
					"Failed to process your request due to an unexpected error. Please try again later.",
					fmt.Errorf("failed to rollback transaction: %w", err),
				)
			}

			return appError.NewAppError(
				http.StatusInternalServerError,
				"Internal Server Error",
				fmt.Sprintf("Failed to send message to WhatsApp: %v", fonnteResp["reason"].(string)),
				fmt.Errorf("failed to send message to WhatsApp: %v", fonnteResp["reason"].(string)),
			)
		}
	}

	return nil
}
