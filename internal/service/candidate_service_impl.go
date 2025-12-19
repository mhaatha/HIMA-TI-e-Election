package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	envConfig "github.com/mhaatha/HIMA-TI-e-Election/internal/config"
	appError "github.com/mhaatha/HIMA-TI-e-Election/internal/errors"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/helper"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/domain"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/web"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/repository"
)

func NewCandidateService(candidateRepository repository.CandidateRepository, envConfig *envConfig.Config, voteService VoteService, db *sql.DB, validate *validator.Validate) CandidateService {
	return &CandidateServiceImpl{
		CandidateRepository: candidateRepository,
		EnvConfig:           envConfig,
		VoteService:         voteService,
		DB:                  db,
		Validate:            validate,
	}
}

type CandidateServiceImpl struct {
	CandidateRepository repository.CandidateRepository
	EnvConfig           *envConfig.Config
	VoteService         VoteService
	DB                  *sql.DB
	Validate            *validator.Validate
}

func (service *CandidateServiceImpl) Create(ctx context.Context, request web.CandidateCreateRequest) (web.CandidateResponse, error) {
	// Validate request
	err := service.Validate.Struct(request)
	if err != nil {
		return web.CandidateResponse{}, appError.NewAppError(
			http.StatusBadRequest,
			"Invalid request payload",
			appError.FormatValidationDetails(err.(validator.ValidationErrors)),
			fmt.Errorf("%w: %v", appError.ErrValidation, err),
		)
	}

	// Open Transaction
	tx, err := service.DB.BeginTx(ctx, nil)
	if err != nil {
		return web.CandidateResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.RollbackQuietly(tx)

	// Get all candidates for the shake of validation
	candidates, err := service.CandidateRepository.GetAll(ctx, tx)
	if err != nil {
		return web.CandidateResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to get all candidates: %w", err),
		)
	}

	// currentPeriod is the current year or period
	currentPeriod := time.Now().Year()

	for _, candidate := range candidates {
		// currentCandidatePeriod is the year or period of the current candidate
		currentCandidatePeriod := candidate.CreatedAt.Year()

		// Numbers cannot be the same in the same period
		if request.Number == candidate.Number && currentPeriod == currentCandidatePeriod {
			return web.CandidateResponse{}, appError.NewAppError(
				http.StatusBadRequest,
				"Invalid request payload",
				"Number is already in use in this period",
				fmt.Errorf("%w", appError.ErrNumberIsUsed),
			)
		}

		// President NIM must be unique
		if request.PresidentNIM == candidate.PresidentNIM || request.PresidentNIM == candidate.ViceNIM {
			return web.CandidateResponse{}, appError.NewAppError(
				http.StatusBadRequest,
				"Invalid request payload",
				"President NIM is already exists",
				fmt.Errorf("%w", appError.ErrNIMAlreadyExists),
			)
		}

		// Vice NIM must be unique
		if request.ViceNIM == candidate.ViceNIM || request.ViceNIM == candidate.PresidentNIM {
			return web.CandidateResponse{}, appError.NewAppError(
				http.StatusBadRequest,
				"Invalid request payload",
				"Vice NIM is already exists",
				fmt.Errorf("%w", appError.ErrNIMAlreadyExists),
			)
		}

		// Photo_key must be unique
		if request.PhotoKey == candidate.PhotoKey {
			return web.CandidateResponse{}, appError.NewAppError(
				http.StatusBadRequest,
				"Invalid request payload",
				fmt.Sprintf("Photo key '%s' is already used, please choose another photo key", request.PhotoKey),
				fmt.Errorf("%w", appError.ErrPhotoKeyIsUsed),
			)
		}
	}

	// Create a variable to store the separated values of mission
	var asterixSeparatedValuesOfMission string
	// Loop through mission and join it with a asterix(*)
	for i, mission := range request.Mission {
		if i == len(request.Mission)-1 {
			asterixSeparatedValuesOfMission += mission
		} else {
			asterixSeparatedValuesOfMission += mission + "*"
		}
	}

	candidate := domain.Candidate{
		Number:                request.Number,
		President:             request.President,
		Vice:                  request.Vice,
		Vision:                request.Vision,
		Mission:               asterixSeparatedValuesOfMission,
		PhotoKey:              request.PhotoKey,
		PresidentStudyProgram: request.PresidentStudyProgram,
		ViceStudyProgram:      request.ViceStudyProgram,
		PresidentNIM:          request.PresidentNIM,
		ViceNIM:               request.ViceNIM,
	}

	// Save candidate to database
	candidate, err = service.CandidateRepository.Save(ctx, tx, candidate)
	if err != nil {
		return web.CandidateResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to create candidate: %w", err),
		)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return web.CandidateResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("transaction commit failed: %w", err),
		)
	}

	// Return candidate with mission as a slice of string for the user
	return helper.ToCandidateResponse(candidate), nil
}

func (service *CandidateServiceImpl) GetCandidates(ctx context.Context, period string) ([]web.CandidateResponseWithURL, error) {
	// Open Transaction
	tx, err := service.DB.BeginTx(ctx, nil)
	if err != nil {
		return []web.CandidateResponseWithURL{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.RollbackQuietly(tx)

	// Create presigned URL for GET
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(service.EnvConfig.S3AccessKeyId, service.EnvConfig.S3SecretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return []web.CandidateResponseWithURL{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrLoadDefaultConfig, err),
		)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(service.EnvConfig.S3URL)
	})

	// Presigned URL client
	presignClient := s3.NewPresignClient(client)

	// Validate query params
	// If period is nil, it means get all candidates
	// If period is not nil, it means get candidates by specific period
	if period != "" {
		// Period must be a positive number and can't exceed 4 digits
		if len(period) > 4 || string(period[0]) == "-" {
			return []web.CandidateResponseWithURL{}, appError.NewAppError(
				http.StatusBadRequest,
				"Invalid request payload",
				"Period must be a positive number and can't exceed 4 digits",
				fmt.Errorf("%w", appError.ErrInvalidPeriodRange),
			)
		}

		periodInt, err := strconv.Atoi(period)
		if err != nil {
			// Period must be a valid number
			if errors.Is(err, strconv.ErrSyntax) {
				return []web.CandidateResponseWithURL{}, appError.NewAppError(
					http.StatusBadRequest,
					"Invalid request payload",
					"Period must be a number and can't contain characters",
					fmt.Errorf("%w: %v", appError.ErrInvalidPeriodSyntax, err),
				)
			}

			return []web.CandidateResponseWithURL{}, appError.NewAppError(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Failed to process your request due to an unexpected error. Please try again later.",
				fmt.Errorf("failed to convert period to int: %v", err),
			)
		}

		// Get candidates by specific period
		candidates, err := service.CandidateRepository.GetByPeriod(ctx, tx, periodInt)
		if err != nil {
			return []web.CandidateResponseWithURL{}, appError.NewAppError(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Failed to process your request due to an unexpected error. Please try again later.",
				fmt.Errorf("failed to get candidate by period: %v", err),
			)
		}

		// Create a place to store candidate with presigned URL
		candidatesWithURL := []domain.CandidateWithURL{}

		for _, candidate := range candidates {
			// Create presigned URL for GetObject in 24 hours
			presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
				Bucket: aws.String(service.EnvConfig.S3Bucket),
				Key:    aws.String(candidate.PhotoKey),
			}, s3.WithPresignExpires(24*time.Hour))
			if err != nil {
				return []web.CandidateResponseWithURL{}, appError.NewAppError(
					http.StatusInternalServerError,
					"Internal Server Error",
					"Failed to process your request due to an unexpected error. Please try again later.",
					fmt.Errorf("%w: %v", appError.ErrCreatePresignedPut, err),
				)
			}

			currentCandidateWithURL := domain.CandidateWithURL{
				Id:                    candidate.Id,
				Number:                candidate.Number,
				President:             candidate.President,
				Vice:                  candidate.Vice,
				Vision:                candidate.Vision,
				Mission:               candidate.Mission,
				PhotoURL:              presignResult.URL,
				PresidentStudyProgram: candidate.PresidentStudyProgram,
				ViceStudyProgram:      candidate.ViceStudyProgram,
				PresidentNIM:          candidate.PresidentNIM,
				ViceNIM:               candidate.ViceNIM,
				CreatedAt:             candidate.CreatedAt,
				UpdatedAt:             candidate.UpdatedAt,
			}

			candidatesWithURL = append(candidatesWithURL, currentCandidateWithURL)
		}

		return helper.ToCandidatesResponseWithURL(candidatesWithURL), nil
	} else {
		// Get all candidates
		candidates, err := service.CandidateRepository.GetAll(ctx, tx)
		if err != nil {
			return []web.CandidateResponseWithURL{}, appError.NewAppError(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Failed to process your request due to an unexpected error. Please try again later.",
				fmt.Errorf("failed to get all candidates: %v", err),
			)
		}

		// Create a place to store candidate with presigned URL
		candidatesWithURL := []domain.CandidateWithURL{}

		for _, candidate := range candidates {
			// Create presigned URL for GetObject in 24 hours
			presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
				Bucket: aws.String(service.EnvConfig.S3Bucket),
				Key:    aws.String(candidate.PhotoKey),
			}, s3.WithPresignExpires(24*time.Hour))
			if err != nil {
				return []web.CandidateResponseWithURL{}, appError.NewAppError(
					http.StatusInternalServerError,
					"Internal Server Error",
					"Failed to process your request due to an unexpected error. Please try again later.",
					fmt.Errorf("%w: %v", appError.ErrCreatePresignedPut, err),
				)
			}

			currentCandidateWithURL := domain.CandidateWithURL{
				Id:                    candidate.Id,
				Number:                candidate.Number,
				President:             candidate.President,
				Vice:                  candidate.Vice,
				Vision:                candidate.Vision,
				Mission:               candidate.Mission,
				PhotoURL:              presignResult.URL,
				PresidentStudyProgram: candidate.PresidentStudyProgram,
				ViceStudyProgram:      candidate.ViceStudyProgram,
				PresidentNIM:          candidate.PresidentNIM,
				ViceNIM:               candidate.ViceNIM,
				CreatedAt:             candidate.CreatedAt,
				UpdatedAt:             candidate.UpdatedAt,
			}

			candidatesWithURL = append(candidatesWithURL, currentCandidateWithURL)
		}

		return helper.ToCandidatesResponseWithURL(candidatesWithURL), nil
	}
}

func (service *CandidateServiceImpl) GetCandidateById(ctx context.Context, candidateId int) (web.CandidateResponseWithURL, error) {
	// Open Transaction
	tx, err := service.DB.BeginTx(ctx, nil)
	if err != nil {
		return web.CandidateResponseWithURL{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.RollbackQuietly(tx)

	// Create presigned URL for GET
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(service.EnvConfig.S3AccessKeyId, service.EnvConfig.S3SecretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return web.CandidateResponseWithURL{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrLoadDefaultConfig, err),
		)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(service.EnvConfig.S3URL)
	})

	// Presigned URL client
	presignClient := s3.NewPresignClient(client)

	// Get candidate by id
	candidate, err := service.CandidateRepository.GetById(ctx, tx, candidateId)
	if err != nil {
		// If candidate not found
		if errors.Is(err, pgx.ErrNoRows) {
			return web.CandidateResponseWithURL{}, appError.NewAppError(
				http.StatusNotFound,
				"Candidate not found",
				fmt.Sprintf("Candidate with id %v does not exist", candidateId),
				fmt.Errorf("candidate with id %v not found: %v", candidateId, err),
			)
		}

		return web.CandidateResponseWithURL{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to get candidate with id %v: %v", candidateId, err),
		)
	}

	// Create presigned URL for GetObject in 24 hours
	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(service.EnvConfig.S3Bucket),
		Key:    aws.String(candidate.PhotoKey),
	}, s3.WithPresignExpires(24*time.Hour))
	if err != nil {
		return web.CandidateResponseWithURL{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrCreatePresignedPut, err),
		)
	}

	candidateWithURL := domain.CandidateWithURL{
		Id:                    candidate.Id,
		Number:                candidate.Number,
		President:             candidate.President,
		Vice:                  candidate.Vice,
		Vision:                candidate.Vision,
		Mission:               candidate.Mission,
		PhotoURL:              presignResult.URL,
		PresidentStudyProgram: candidate.PresidentStudyProgram,
		ViceStudyProgram:      candidate.ViceStudyProgram,
		PresidentNIM:          candidate.PresidentNIM,
		ViceNIM:               candidate.ViceNIM,
		CreatedAt:             candidate.CreatedAt,
		UpdatedAt:             candidate.UpdatedAt,
	}

	return helper.ToCandidateResponseWithURL(candidateWithURL), nil
}

func (service *CandidateServiceImpl) UpdateCandidateById(ctx context.Context, candidateId int, request web.CandidateUpdateRequest) (web.CandidateResponse, error) {
	// Validate request
	err := service.Validate.Struct(request)
	if err != nil {
		return web.CandidateResponse{}, appError.NewAppError(
			http.StatusBadRequest,
			"Invalid request payload",
			appError.FormatValidationDetails(err.(validator.ValidationErrors)),
			fmt.Errorf("failed to update candidate with id %v: %v", candidateId, err),
		)
	}

	// Open Transaction
	tx, err := service.DB.BeginTx(ctx, nil)
	if err != nil {
		return web.CandidateResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.RollbackQuietly(tx)

	// Request body cannot be empty validation
	if helper.IsEmptyStruct(request) {
		return web.CandidateResponse{}, appError.NewAppError(
			http.StatusBadRequest,
			"Invalid request payload",
			"Request body cannot be empty",
			fmt.Errorf("failed to get candidate with id %v: %v", candidateId, err),
		)
	}

	// Is candidateId exists in database
	candidate, err := service.CandidateRepository.GetById(ctx, tx, candidateId)
	if err != nil {
		// If candidate not found
		if errors.Is(err, pgx.ErrNoRows) {
			return web.CandidateResponse{}, appError.NewAppError(
				http.StatusNotFound,
				"Candidate not found",
				fmt.Sprintf("Candidate with id '%v' does not exist", candidateId),
				fmt.Errorf("candidate with id %v not found: %v", candidateId, err),
			)
		}

		return web.CandidateResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to get candidate with id %v: %v", candidateId, err),
		)
	}

	// Check whether candidate's record is already on votes table
	votes, err := service.VoteService.GetByCandidateId(ctx, candidate.Id)
	if err != nil {
		return web.CandidateResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to get votes by candidate id: %v", err),
		)
	}

	// If votes has data, candidate cannot be deleted
	if len(votes) > 0 {
		return web.CandidateResponse{}, appError.NewAppError(
			http.StatusConflict,
			"Candidate has votes",
			"Cannot delete candidate because votes have already submitted",
			fmt.Errorf("%w", appError.ErrCandidateHasVotes),
		)
	}

	// Check the request body, if exists, swap the candidate to the request body
	if request.Number != 0 {
		candidate.Number = request.Number
	}
	if request.President != "" {
		candidate.President = request.President
	}
	if request.Vice != "" {
		candidate.Vice = request.Vice
	}
	if request.Vision != "" {
		candidate.Vision = request.Vision
	}
	if request.Mission != nil {
		candidate.Mission = strings.Join(request.Mission, "*")
	}
	if request.PhotoKey != "" {
		candidate.PhotoKey = request.PhotoKey
	}
	if request.PresidentStudyProgram != "" {
		candidate.PresidentStudyProgram = request.PresidentStudyProgram
	}
	if request.ViceStudyProgram != "" {
		candidate.ViceStudyProgram = request.ViceStudyProgram
	}
	if request.PresidentNIM != "" {
		candidate.PresidentNIM = request.PresidentNIM
	}
	if request.ViceNIM != "" {
		candidate.ViceNIM = request.ViceNIM
	}

	// Get all candidates for the shake of validation
	candidates, err := service.CandidateRepository.GetAll(ctx, tx)
	if err != nil {
		return web.CandidateResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to get all candidates: %w", err),
		)
	}

	// currentPeriod is the current year or period
	currentPeriod := time.Now().Year()

	for _, c := range candidates {
		if c.Id == candidateId {
			continue
		}

		// currentCandidatePeriod is the year or period of the current candidate
		currentCandidatePeriod := c.CreatedAt.Year()

		// Numbers cannot be the same in the same period
		if request.Number == c.Number && currentPeriod == currentCandidatePeriod {
			return web.CandidateResponse{}, appError.NewAppError(
				http.StatusBadRequest,
				"Invalid request payload",
				"Number is already in use in this period",
				fmt.Errorf("%w", appError.ErrNumberIsUsed),
			)
		}

		// President NIM must be unique
		if request.PresidentNIM == c.PresidentNIM || request.PresidentNIM == c.ViceNIM {
			return web.CandidateResponse{}, appError.NewAppError(
				http.StatusBadRequest,
				"Invalid request payload",
				"President NIM is already exists",
				fmt.Errorf("%w", appError.ErrNIMAlreadyExists),
			)
		}

		// Vice NIM must be unique
		if request.ViceNIM == c.ViceNIM || request.ViceNIM == c.PresidentNIM {
			return web.CandidateResponse{}, appError.NewAppError(
				http.StatusBadRequest,
				"Invalid request payload",
				"Vice NIM is already exists",
				fmt.Errorf("%w", appError.ErrNIMAlreadyExists),
			)
		}

		// Photo_key must be unique
		if request.PhotoKey == c.PhotoKey {
			return web.CandidateResponse{}, appError.NewAppError(
				http.StatusBadRequest,
				"Invalid request payload",
				fmt.Sprintf("Photo key '%s' is already used, please choose another photo key", request.PhotoKey),
				fmt.Errorf("%w", appError.ErrPhotoKeyIsUsed),
			)
		}
	}

	// Call update repository
	candidate, err = service.CandidateRepository.UpdateById(ctx, tx, candidateId, candidate)
	if err != nil {
		return web.CandidateResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to update candidate with id %v: %v", candidateId, err),
		)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return web.CandidateResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("transaction commit failed: %w", err),
		)
	}

	return helper.ToCandidateResponse(candidate), nil
}

func (service *CandidateServiceImpl) DeleteCandidateById(ctx context.Context, candidateId int) error {
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

	// Get candidate by id
	candidate, err := service.CandidateRepository.GetById(ctx, tx, candidateId)
	if err != nil {
		// If candidate not found
		if errors.Is(err, pgx.ErrNoRows) {
			return appError.NewAppError(
				http.StatusNotFound,
				"Candidate not found",
				fmt.Sprintf("Candidate with id %v does not exist", candidateId),
				fmt.Errorf("candidate with id %v not found: %v", candidateId, err),
			)
		}

		return appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to get candidate with id %v: %v", candidateId, err),
		)
	}

	// Check whether candidate's record is already on votes table
	votes, err := service.VoteService.GetByCandidateId(ctx, candidate.Id)
	if err != nil {
		return appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to get votes by candidate id: %v", err),
		)
	}

	// If votes has data, candidate cannot be deleted
	if len(votes) > 0 {
		return appError.NewAppError(
			http.StatusConflict,
			"Candidate has votes",
			"Cannot delete candidate because votes have already submitted",
			fmt.Errorf("%w", appError.ErrCandidateHasVotes),
		)
	}

	// Setup S3
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(service.EnvConfig.S3AccessKeyId, service.EnvConfig.S3SecretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrLoadDefaultConfig, err),
		)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(service.EnvConfig.S3URL)
	})

	// Delete candidate photo
	_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(service.EnvConfig.S3Bucket),
		Key:    aws.String(candidate.PhotoKey),
	})
	if err != nil {
		return appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrCreatePresignedPut, err),
		)
	}

	// Delete candidate by id
	err = service.CandidateRepository.DeleteById(ctx, tx, candidateId)
	if err != nil {
		return appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to delete candidate with id %v: %v", candidateId, err),
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
