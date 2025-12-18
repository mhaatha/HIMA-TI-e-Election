package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mhaatha/HIMA-TI-e-Election/config"
	appError "github.com/mhaatha/HIMA-TI-e-Election/errors"
	"github.com/mhaatha/HIMA-TI-e-Election/helper"
	"github.com/mhaatha/HIMA-TI-e-Election/model/domain"
	"github.com/mhaatha/HIMA-TI-e-Election/model/web"
	"github.com/mhaatha/HIMA-TI-e-Election/repository"
)

func NewVoteService(voteRepository repository.VoteRepository, votingAccessRepository repository.VotingAccessRepository, authRepository repository.AuthRepository, userService UserService, db *pgxpool.Pool, validate *validator.Validate) VoteService {
	return &VoteServiceImpl{
		VoteRepository:         voteRepository,
		VotingAccessRepository: votingAccessRepository,
		AuthRepository:         authRepository,
		UserService:            userService,
		DB:                     db,
		Validate:               validate,
	}
}

type VoteServiceImpl struct {
	VoteRepository         repository.VoteRepository
	CandidateService       CandidateService
	VotingAccessRepository repository.VotingAccessRepository
	AuthRepository         repository.AuthRepository
	UserService            UserService
	DB                     *pgxpool.Pool
	Validate               *validator.Validate
}

// Setter to fix Circular Dependency with candidateService
func (service *VoteServiceImpl) SetCandidateService(candidateService CandidateService) {
	service.CandidateService = candidateService
}

func (service *VoteServiceImpl) GetByCandidateId(ctx context.Context, candidateId int) ([]web.VoteResponse, error) {
	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return []web.VoteResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Call repository
	votes, err := service.VoteRepository.GetByCandidateId(ctx, tx, candidateId)
	if err != nil {
		return []web.VoteResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to get votes by candidate id: %v", err),
		)
	}

	return helper.ToVotesResponse(votes), nil
}

func (service *VoteServiceImpl) SaveVoteRecord(ctx context.Context, request web.VoteCreateRequest, userId int) (web.VoteCreateResponse, error) {
	// Validate request
	err := service.Validate.Struct(request)
	if err != nil {
		return web.VoteCreateResponse{}, appError.NewAppError(
			http.StatusBadRequest,
			"Invalid request payload",
			appError.FormatValidationDetails(err.(validator.ValidationErrors)),
			fmt.Errorf("%w: %v", appError.ErrValidation, err),
		)
	}

	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return web.VoteCreateResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Only candidate in this period that can be voted
	// Get candidate by candidate_id
	candidate, err := service.CandidateService.GetCandidateById(ctx, request.CandidateId)
	if err != nil {
		return web.VoteCreateResponse{}, err
	}

	// Check if candidate is in this period
	currentPeriod := time.Now().Year()
	if candidate.CreatedAt.Year() != currentPeriod {
		return web.VoteCreateResponse{}, appError.NewAppError(
			http.StatusNotFound,
			"Candidate not found",
			fmt.Sprintf("Candidate with id %v does not exist", request.CandidateId),
			fmt.Errorf("candidate not in this period: %v", err),
		)
	}

	// Users can only vote once in a period. Users can vote again if the period has changed.
	votingAccess, err := service.VotingAccessRepository.GetByUserId(ctx, tx, userId)
	if err != nil {
		return web.VoteCreateResponse{}, appError.NewAppError(
			http.StatusForbidden,
			"Forbidden",
			"You don't have an access to vote",
			fmt.Errorf("failed to get voting access by user id: %v", err),
		)
	}
	// Check if user has voted in this period
	exists, err := service.VoteRepository.IsUserVotedInPeriod(ctx, tx, votingAccess.Hashed)
	if err != nil {
		return web.VoteCreateResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to check if user has voted in this period: %v", err),
		)
	}
	if exists {
		return web.VoteCreateResponse{}, appError.NewAppError(
			http.StatusBadRequest,
			"User has already voted",
			"User has already voted in this period",
			fmt.Errorf("user has already voted in this period: %v", err),
		)
	}

	// Call repository
	vote, err := service.VoteRepository.SaveVoteRecord(ctx, tx, domain.Vote{
		CandidateId: request.CandidateId,
		HashedNim:   votingAccess.Hashed,
	})
	if err != nil {
		return web.VoteCreateResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to save vote record: %v", err),
		)
	}

	// Get user by user_id
	user, err := service.UserService.GetById(ctx, userId)
	if err != nil {
		return web.VoteCreateResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to get user by user_id: %v", err),
		)
	}

	// Write the log to the file
	config.FileLog.Infof("%s with NIM %s from %s has voted", user.FullName, user.NIM, user.StudyProgram)

	return web.VoteCreateResponse{
		CreatedAt: vote.CreatedAt,
	}, nil
}

func (service *VoteServiceImpl) GetTotalVotesByCandidateId(ctx context.Context, candidateId int) (web.TotalVoteResponse, error) {
	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return web.TotalVoteResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Call repository
	totalVotes, err := service.VoteRepository.GetTotalVotesByCandidateId(ctx, tx, candidateId)
	if err != nil {
		// If candidate is not found
		if errors.Is(err, pgx.ErrNoRows) {
			return web.TotalVoteResponse{}, appError.NewAppError(
				http.StatusNotFound,
				"Candidate not found",
				fmt.Sprintf("Candidate with id %v does not exist", candidateId),
				fmt.Errorf("candidate with id %v not found: %v", candidateId, err),
			)
		}

		return web.TotalVoteResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to get total votes by candidate id: %v", err),
		)
	}

	return web.TotalVoteResponse{
		TotalVotes: totalVotes,
	}, nil
}

func (service *VoteServiceImpl) StreamVoteEvents(ctx context.Context, conn *websocket.Conn) {
	dbConn, err := service.DB.Acquire(context.Background())
	if err != nil {
		appError.LogError(err, "failed to acquire DB")
		return
	}
	defer dbConn.Release()

	config.Log.Info("listening to votes_channel...")

	_, err = dbConn.Exec(context.Background(), "LISTEN votes_channel")
	if err != nil {
		appError.LogError(err, "failed to LISTEN")
		return
	}

	for {
		select {
		case <-ctx.Done():
			config.Log.Info("Client disconnected, stopping stream...")
			return
		default:
			notifications, err := dbConn.Conn().WaitForNotification(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				appError.LogError(err, "WaitForNotification error")
				continue
			}
			config.Log.Info(fmt.Sprintf("Received notifications: %v", notifications.Payload))

			if err := conn.WriteMessage(websocket.TextMessage, []byte(notifications.Payload)); err != nil {
				appError.LogError(err, "failed to write message")
				return
			}
		}
	}
}

func (service *VoteServiceImpl) CheckIfUserHasVoted(ctx context.Context, sessionId string) (bool, error) {
	// Open Transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return false, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrTransaction, err),
		)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Check if sessionId is not from admin
	// Get session by sessionId
	session, err := service.AuthRepository.GetSessionById(ctx, tx, sessionId)
	if err != nil {
		return false, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to get session by sessionId: %v", err),
		)
	}

	// Get user by userId
	user, err := service.UserService.GetById(ctx, session.UserId)
	if err != nil {
		return false, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to get user by userId: %v", err),
		)
	}

	if user.Role == "admin" {
		return false, appError.NewAppError(
			http.StatusBadRequest,
			"Admin cannot vote",
			"Admin does not have permission to vote",
			fmt.Errorf("%s is an admin, cannot vote", user.FullName),
		)
	}

	// Check is user already voted in this period
	isVoted, err := service.VoteRepository.IsUserEverVoted(ctx, tx, session.UserId)
	if err != nil {
		return false, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("failed to check if user has ever voted: %v", err),
		)
	}

	return isVoted, nil
}
