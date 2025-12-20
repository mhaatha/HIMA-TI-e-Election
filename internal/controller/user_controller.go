package controller

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type UserController interface {
	Create(w http.ResponseWriter, r *http.Request, params httprouter.Params)
	UpdateCurrent(w http.ResponseWriter, r *http.Request, params httprouter.Params)
	GetCurrent(w http.ResponseWriter, r *http.Request, params httprouter.Params)
	GetAll(w http.ResponseWriter, r *http.Request, params httprouter.Params)
	UpdateById(w http.ResponseWriter, r *http.Request, params httprouter.Params)
	DeleteById(w http.ResponseWriter, r *http.Request, params httprouter.Params)
	BulkCreate(w http.ResponseWriter, r *http.Request, params httprouter.Params)
	GeneratePassword(w http.ResponseWriter, r *http.Request, params httprouter.Params)
}
