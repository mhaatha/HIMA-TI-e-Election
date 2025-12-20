package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	appError "github.com/mhaatha/HIMA-TI-e-Election/internal/errors"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/helper"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/web"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/service"
)

type contextKey string

const SessionContextKey contextKey = "session"

func UserMiddleware(next httprouter.Handle, authService service.AuthService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Get cookie from header
		cookie, err := r.Cookie("e_election_session")

		// Check if cookie exists
		if err != nil || cookie.Value == "" {
			appError.LogError(err, "cookie not found")

			w.WriteHeader(http.StatusUnauthorized)
			helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
				"error": {
					Message: "Cookie not found",
					Details: "Either cookie not found or cookie value is empty",
				},
			})
			return
		}

		// Call service
		session, err := authService.UserValidateSession(r.Context(), cookie.Value)
		if err != nil {
			var customError *appError.AppError

			if errors.As(err, &customError) {
				appError.LogError(err, "invalid session")

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

		// Send session data through context
		ctx := context.WithValue(r.Context(), SessionContextKey, session)
		next(w, r.WithContext(ctx), params)
	}
}

func AdminMiddleware(next httprouter.Handle, authService service.AuthService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Get cookie from header
		cookie, err := r.Cookie("e_election_session")

		// Check if cookie exists
		if err != nil || cookie.Value == "" {
			appError.LogError(err, "cookie not found")

			w.WriteHeader(http.StatusUnauthorized)
			helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
				"error": {
					Message: "Cookie not found",
					Details: "Either cookie not found or cookie value is empty",
				},
			})
			return
		}

		// Call service
		session, err := authService.AdminValidateSession(r.Context(), cookie.Value)
		if err != nil {
			var customError *appError.AppError

			if errors.As(err, &customError) {
				appError.LogError(err, "invalid session")

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

		// Send session data through context
		ctx := context.WithValue(r.Context(), SessionContextKey, session)
		next(w, r.WithContext(ctx), params)
	}
}
