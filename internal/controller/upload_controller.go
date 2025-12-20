package controller

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type UploadController interface {
	GetPresignedUrl(w http.ResponseWriter, r *http.Request, params httprouter.Params)
}
