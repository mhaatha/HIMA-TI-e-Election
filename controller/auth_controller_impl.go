package controller

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	appError "github.com/mhaatha/HIMA-TI-e-Election/errors"
	"github.com/mhaatha/HIMA-TI-e-Election/helper"
	"github.com/mhaatha/HIMA-TI-e-Election/middleware"
	"github.com/mhaatha/HIMA-TI-e-Election/model/web"
	"github.com/mhaatha/HIMA-TI-e-Election/service"
)

const (
	MaxAge      = 86400 // 24 Hour
	SessionName = "e_election_session"
)

func NewAuthController(authService service.AuthService) AuthController {
	return &AuthControllerImpl{
		AuthService: authService,
	}
}

type AuthControllerImpl struct {
	AuthService service.AuthService
}

func (controller *AuthControllerImpl) Login(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Get request body and write it to loginRequest
	loginRequest := web.LoginRequest{}
	err := helper.ReadFromRequestBody(r, &loginRequest)
	if err != nil {
		helper.WriteJSONDecodeError(w, err)
		return
	}

	var loginResponse web.LoginResponse
	var sessionData string

	// If NIM is empty that means login as admin
	if loginRequest.NIM == "" {
		loginResponse, sessionData, err = controller.AuthService.LoginAdmin(r.Context(), MaxAge, loginRequest)
	} else {
		loginResponse, sessionData, err = controller.AuthService.LoginUser(r.Context(), MaxAge, loginRequest)
	}

	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(err, "login failed")

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

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SessionName,
		Value:    sessionData,
		MaxAge:   MaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		Path:     "/",
	})

	// Send response
	w.WriteHeader(http.StatusOK)
	helper.WriteToResponseBody(w, web.WebSuccessResponse{
		Message: "Login success",
		Data:    loginResponse,
	})
}

func (controller *AuthControllerImpl) Logout(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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
	err := controller.AuthService.Logout(r.Context(), cookie.SessionId)
	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(customError, "logout failed")

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

	// Delete cookie in browser
	http.SetCookie(w, &http.Cookie{
		Name:     SessionName,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})

	// Send response
	w.WriteHeader(http.StatusOK)
	helper.WriteToResponseBody(w, web.WebSuccessResponse{
		Message: "Logout success",
		Data:    nil,
	})
}
