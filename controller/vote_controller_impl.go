package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
	"github.com/mhaatha/HIMA-TI-e-Election/config"
	myErrors "github.com/mhaatha/HIMA-TI-e-Election/errors"
	"github.com/mhaatha/HIMA-TI-e-Election/helper"
	"github.com/mhaatha/HIMA-TI-e-Election/middleware"
	"github.com/mhaatha/HIMA-TI-e-Election/model/web"
	"github.com/mhaatha/HIMA-TI-e-Election/service"
)

var (
	clients        = make(map[*websocket.Conn]bool)
	broadcast      = make(chan string)
	mutex          sync.Mutex
	feFirstDomain  = "http://localhost:5500"
	feSecondDomain = "http://127.0.0.1:5500"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")

		return origin == feFirstDomain || origin == feSecondDomain
	},
}

func NewVoteController(voteService service.VoteService, db *pgxpool.Pool) VoteController {
	return &VoteControllerImpl{
		VoteService: voteService,
		DB:          db,
	}
}

type VoteControllerImpl struct {
	VoteService service.VoteService
	DB          *pgxpool.Pool
}

func (controller *VoteControllerImpl) Save(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Get cookie from context
	cookie, ok := r.Context().Value(middleware.SessionContextKey).(web.SessionResponse)
	if !ok {
		myErrors.LogError(nil, "invalid session data")

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
		var customError *myErrors.AppError

		if errors.As(err, &customError) {
			myErrors.LogError(err, "failed to save vote record")

			w.WriteHeader(customError.StatusCode)
			helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
				"error": {
					Message: customError.Message,
					Details: customError.Details,
				},
			})
			return
		} else {
			myErrors.LogError(err, "unexpected error")

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
		var customError *myErrors.AppError

		if errors.As(err, &customError) {
			myErrors.LogError(err, "failed to get total votes")

			w.WriteHeader(customError.StatusCode)
			helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
				"error": {
					Message: customError.Message,
					Details: customError.Details,
				},
			})
			return
		} else {
			myErrors.LogError(err, "unexpected error")

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
		myErrors.LogError(err, "failed to upgrade connection")

		w.WriteHeader(http.StatusInternalServerError)
		helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
			"error": {
				Message: "Internal Server Error",
				Details: "Internal Server Error. Please try again later.",
			},
		})
		return
	}

	mutex.Lock()
	clients[conn] = true
	mutex.Unlock()

	for msg := range broadcast {
		mutex.Lock()
		for client := range clients {
			if err := client.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}

	controller.VoteService.StreamVoteEvents(r.Context(), conn)
}

func (controller *VoteControllerImpl) ListenToDB(ctx context.Context) {
	conn, err := controller.DB.Acquire(ctx)
	if err != nil {
		panic(err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "LISTEN votes_channel")
	if err != nil {
		panic(err)
	}

	config.Log.Info("listening to votes_channel...")

	for {
		notif, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			continue
		}
		broadcast <- notif.Payload
	}
}
