package controller

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type AuthController interface {
	Login(w http.ResponseWriter, r *http.Request, params httprouter.Params)
	Logout(w http.ResponseWriter, r *http.Request, params httprouter.Params)
}
