package controller

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	myErrors "github.com/mhaatha/HIMA-TI-e-Election/errors"
	"github.com/mhaatha/HIMA-TI-e-Election/helper"
	"github.com/mhaatha/HIMA-TI-e-Election/model/web"
	"github.com/mhaatha/HIMA-TI-e-Election/service"
)

func NewUploadController(uploadService service.UploadService) UploadController {
	return &UploadControllerImpl{
		UploadService: uploadService,
	}
}

type UploadControllerImpl struct {
	UploadService service.UploadService
}

func (controller *UploadControllerImpl) GetPresignedUrl(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Get query paramaters
	filename := r.URL.Query().Get("filename")

	if filename != "" {
		data, err := controller.UploadService.CreatePresignedURL(r.Context(), filename)
		if err != nil {
			var customError *myErrors.AppError

			if errors.As(err, &customError) {
				myErrors.LogError(err, "failed to get presigned URL")

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
			Message: "Success get presigned url",
			Data:    data,
		})
	} else {
		w.WriteHeader(http.StatusBadRequest)
		helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
			"error": {
				Message: "Invalid request query parameters",
				Details: "Query parameter 'filename' is required",
			},
		})
		return
	}
}
