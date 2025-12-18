package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	appError "github.com/mhaatha/HIMA-TI-e-Election/errors"
	"github.com/mhaatha/HIMA-TI-e-Election/helper"
	"github.com/mhaatha/HIMA-TI-e-Election/model/web"
	"github.com/mhaatha/HIMA-TI-e-Election/service"
)

func NewCandidateController(candidateService service.CandidateService) CandidateController {
	return &CandidateControllerImpl{
		CandidateService: candidateService,
	}
}

type CandidateControllerImpl struct {
	CandidateService service.CandidateService
}

func (controller *CandidateControllerImpl) Create(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Get request body and write it to candidateRequest
	candidateRequest := web.CandidateCreateRequest{}
	err := helper.ReadFromRequestBody(r, &candidateRequest)
	if err != nil {
		helper.WriteJSONDecodeError(w, err)
		return
	}

	// Call service
	candidateResponse, err := controller.CandidateService.Create(r.Context(), candidateRequest)
	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(err, "failed to create candidate")

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

	// Write and send the response
	w.WriteHeader(http.StatusOK)
	helper.WriteToResponseBody(w, web.WebSuccessResponse{
		Message: "Success create candidate",
		Data:    candidateResponse,
	})
}

func (controller *CandidateControllerImpl) GetCandidates(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Get query paramaters
	period := r.URL.Query().Get("period")

	// Call service
	candidates, err := controller.CandidateService.GetCandidates(r.Context(), period)
	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(err, "failed to get candidates")

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

	// Write and send the response
	w.WriteHeader(http.StatusOK)
	helper.WriteToResponseBody(w, web.WebSuccessResponse{
		Message: "Success get candidates",
		Data:    candidates,
	})
}

func (controller *CandidateControllerImpl) GetCandidateById(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Get id from named parameter
	candidateId := params.ByName("candidateId")

	// Convert query params to int
	candidateIdInt, err := strconv.Atoi(candidateId)
	if err != nil {
		appError.LogError(err, "failed to convert id to int")

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
	candidate, err := controller.CandidateService.GetCandidateById(r.Context(), candidateIdInt)
	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(err, "failed to get candidate")

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

	// Write and send the response
	w.WriteHeader(http.StatusOK)
	helper.WriteToResponseBody(w, web.WebSuccessResponse{
		Message: "Success get candidate",
		Data:    candidate,
	})
}

func (controller *CandidateControllerImpl) UpdateCandidateById(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Get id from named parameter
	candidateId := params.ByName("candidateId")

	// Convert query params to int
	candidateIdInt, err := strconv.Atoi(candidateId)
	if err != nil {
		appError.LogError(err, "failed to convert id to int")

		w.WriteHeader(http.StatusNotFound)
		helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
			"error": {
				Message: "Candidate not found",
				Details: fmt.Sprintf("Candidate with id '%v' does not exist", candidateId),
			},
		})
		return
	}

	// Get request body and write it to candidateRequest
	candidateRequest := web.CandidateUpdateRequest{}
	err = helper.ReadFromRequestBody(r, &candidateRequest)
	if err != nil {
		helper.WriteJSONDecodeError(w, err)
		return
	}

	// Call service
	updatedCandidate, err := controller.CandidateService.UpdateCandidateById(r.Context(), candidateIdInt, candidateRequest)
	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(err, "failed to update candidate")

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

	// Write and send the response
	w.WriteHeader(http.StatusOK)
	helper.WriteToResponseBody(w, web.WebSuccessResponse{
		Message: "Success update candidate",
		Data:    updatedCandidate,
	})
}

func (controller *CandidateControllerImpl) DeleteCandidateById(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Get id from named parameter
	candidateId := params.ByName("candidateId")

	// Convert query params to int
	candidateIdInt, err := strconv.Atoi(candidateId)
	if err != nil {
		appError.LogError(err, "failed to convert id to int")

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
	err = controller.CandidateService.DeleteCandidateById(r.Context(), candidateIdInt)
	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(err, "failed to delete candidate")

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

	// Write and send the response
	w.WriteHeader(http.StatusOK)
	helper.WriteToResponseBody(w, web.WebSuccessResponse{
		Message: "Success delete candidate",
		Data:    nil,
	})
}
