package middleware

import (
	"net/http"

	"github.com/mhaatha/HIMA-TI-e-Election/internal/config"
	"github.com/sirupsen/logrus"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			config.Log.WithFields(logrus.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
			}).Info("request received from client")

			next.ServeHTTP(w, r)
		},
	)
}
