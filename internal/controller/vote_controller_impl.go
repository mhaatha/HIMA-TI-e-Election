package controller

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/julienschmidt/httprouter"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/config"
	appError "github.com/mhaatha/HIMA-TI-e-Election/internal/errors"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/helper"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/middleware"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/web"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/service"
)

var (
	feFirstDomain  = "http://localhost:5500"
	feSecondDomain = "https://himati-e-election-polnes.vercel.app"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")

		return origin == feFirstDomain || origin == feSecondDomain
	},
}

func NewVoteController(voteService service.VoteService, db *sql.DB) VoteController {
	return &VoteControllerImpl{
		VoteService:  voteService,
		DB:           db,
		Clients:      make(map[chan string]bool),
		ClientsMutex: &sync.Mutex{},
	}
}

type VoteControllerImpl struct {
	VoteService service.VoteService
	DB          *sql.DB

	Clients      map[chan string]bool
	ClientsMutex *sync.Mutex
}

// Helper to register new client
func (controller *VoteControllerImpl) AddClient(clientChan chan string) {
	controller.ClientsMutex.Lock()
	defer controller.ClientsMutex.Unlock()

	controller.Clients[clientChan] = true
}

// Helper to delete disconnected client
func (controller *VoteControllerImpl) RemoveClient(clientChan chan string) {
	controller.ClientsMutex.Lock()
	defer controller.ClientsMutex.Unlock()

	delete(controller.Clients, clientChan)
	close(clientChan)
}

// Helper to broadcast the message to all clients
func (controller *VoteControllerImpl) BroadcastToClients(message string) {
	controller.ClientsMutex.Lock()
	defer controller.ClientsMutex.Unlock()

	for clientChan := range controller.Clients {
		select {
		case clientChan <- message:
		default:
			// Optional: Kick client if the buffer is full
		}
	}
}

func (controller *VoteControllerImpl) Save(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Get cookie from context
	cookie, ok := r.Context().Value(middleware.SessionContextKey).(web.SessionResponse)
	if !ok {
		appError.LogError(nil, "invalid session data")

		w.WriteHeader(http.StatusUnauthorized)
		helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
			"error": {
				Message: "Invalid session data",
				Details: "Session data may be corrupted or missing",
			},
		})
		return
	}

	// Get request body and write it to voteRequest
	voteRequest := web.VoteCreateRequest{}
	err := helper.ReadFromRequestBody(r, &voteRequest)
	if err != nil {
		helper.WriteJSONDecodeError(w, err)
		return
	}

	// Call service
	voteResponse, err := controller.VoteService.SaveVoteRecord(r.Context(), voteRequest, cookie.UserId)
	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(err, "failed to save vote record")

			w.WriteHeader(customError.StatusCode)
			helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
				"error": {
					Message: customError.Message,
					Details: customError.Details,
				},
			})
			return
		} else {
			appError.LogError(err, "unexpected error")

			w.WriteHeader(http.StatusInternalServerError)
			helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
				"error": {
					Message: "Internal Server Error",
					Details: "Internal Server Error. Please try again later.",
				},
			})
			return
		}
	}

	// Write and send the response
	w.WriteHeader(http.StatusOK)
	helper.WriteToResponseBody(w, web.WebSuccessResponse{
		Message: "Vote has been recorded successfully",
		Data:    voteResponse,
	})
}

func (controller *VoteControllerImpl) GetTotalVotesByCandidateId(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Get id from named parameter
	candidateId := params.ByName("candidateId")

	// Convert query params to int
	candidateIdInt, err := strconv.Atoi(candidateId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
			"error": {
				Message: "Candidate not found",
				Details: fmt.Sprintf("Candidate with id '%v' does not exist", candidateId),
			},
		})
		return
	}

	// Call service
	totalVotes, err := controller.VoteService.GetTotalVotesByCandidateId(r.Context(), candidateIdInt)
	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(err, "failed to get total votes")

			w.WriteHeader(customError.StatusCode)
			helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
				"error": {
					Message: customError.Message,
					Details: customError.Details,
				},
			})
			return
		} else {
			appError.LogError(err, "unexpected error")

			w.WriteHeader(http.StatusInternalServerError)
			helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
				"error": {
					Message: "Internal Server Error",
					Details: "Internal Server Error. Please try again later.",
				},
			})
			return
		}
	}

	// Write and send the response
	w.WriteHeader(http.StatusOK)
	helper.WriteToResponseBody(w, web.WebSuccessResponse{
		Message: "Success get total votes by candidate id",
		Data:    totalVotes,
	})
}

func (controller *VoteControllerImpl) VotesLiveResult(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		appError.LogError(err, "failed to upgrade connection")
		return
	}

	controller.StreamVoteEvents(r.Context(), conn)
}

func (controller *VoteControllerImpl) ListenToDB(ctx context.Context) {
	for {
		conn, err := controller.DB.Conn(ctx)
		if err != nil {
			config.Log.Error("failed to acquire connection from pool:", err)
			time.Sleep(5 * time.Second) // Retry delay
			continue
		}

		err = conn.Raw(func(driverConn any) error {
			stdlibConn, ok := driverConn.(*stdlib.Conn)
			if !ok {
				return fmt.Errorf("connection is not a pgx connection")
			}

			pgxConn := stdlibConn.Conn()

			config.Log.Info("Starting LISTEN on votes_channel...")

			_, err := pgxConn.Exec(ctx, "LISTEN votes_channel")
			if err != nil {
				return err
			}

			// Loop Notification
			for {
				if ctx.Err() != nil {
					return ctx.Err()
				}

				// WaitForNotification (Blocking)
				notification, err := pgxConn.WaitForNotification(ctx)
				if err != nil {
					return err
				}

				// Broadcast to Hub
				controller.BroadcastToClients(notification.Payload)
			}
		})

		if err != nil {
			config.Log.Errorf("listener disconnected: %v. reconnecting...", err)
		}

		conn.Close()

		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
			continue
		}
	}
}

func (controller *VoteControllerImpl) CheckIfUserHasVoted(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Get cookie from context
	cookie, ok := r.Context().Value(middleware.SessionContextKey).(web.SessionResponse)
	if !ok {
		appError.LogError(nil, "invalid session data")

		w.WriteHeader(http.StatusUnauthorized)
		helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
			"error": {
				Message: "Invalid session data",
				Details: "Session data may be corrupted or missing",
			},
		})
		return
	}

	// Call service
	isVoted, err := controller.VoteService.CheckIfUserHasVoted(r.Context(), cookie.SessionId)
	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(err, "failed to check if user has voted")

			w.WriteHeader(customError.StatusCode)
			helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
				"error": {
					Message: customError.Message,
					Details: customError.Details,
				},
			})
			return
		} else {
			appError.LogError(err, "unexpected error")

			w.WriteHeader(http.StatusInternalServerError)
			helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
				"error": {
					Message: "Internal Server Error",
					Details: "Internal Server Error. Please try again later.",
				},
			})
			return
		}
	}

	// Write and send the response
	w.WriteHeader(http.StatusOK)
	helper.WriteToResponseBody(w, web.WebSuccessResponse{
		Message: "Success check if user has voted",
		Data:    map[string]bool{"has_voted": isVoted},
	})
}

func (controller *VoteControllerImpl) StreamVoteEvents(ctx context.Context, wsConn *websocket.Conn) {
	clientChan := make(chan string, 10)

	controller.AddClient(clientChan)
	config.Log.Info("new WebSocket client subscribed to vote events")

	defer func() {
		controller.RemoveClient(clientChan)
		config.Log.Info("webSocket client unsubscribed")

		wsConn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return

		case payload, ok := <-clientChan:
			if !ok {
				return
			}

			if err := wsConn.WriteMessage(websocket.TextMessage, []byte(payload)); err != nil {
				appError.LogError(err, "failed to write to websocket")
				return
			}
		}
	}
}
