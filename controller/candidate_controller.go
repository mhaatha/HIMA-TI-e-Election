package controller

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type CandidateController interface {
	Create(w http.ResponseWriter, r *http.Request, params httprouter.Params)
	GetCandidates(w http.ResponseWriter, r *http.Request, params httprouter.Params)
	GetCandidateById(w http.ResponseWriter, r *http.Request, params httprouter.Params)
	UpdateCandidateById(w http.ResponseWriter, r *http.Request, params httprouter.Params)
	DeleteCandidateById(w http.ResponseWriter, r *http.Request, params httprouter.Params)
}
