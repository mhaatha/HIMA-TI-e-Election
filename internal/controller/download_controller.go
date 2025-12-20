package controller

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type DownloadController interface {
	GetPresignedUrl(w http.ResponseWriter, r *http.Request, params httprouter.Params)
}
