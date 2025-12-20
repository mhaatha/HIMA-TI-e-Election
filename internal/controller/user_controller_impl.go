package controller

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	appError "github.com/mhaatha/HIMA-TI-e-Election/internal/errors"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/helper"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/middleware"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/web"
	"github.com/mhaatha/HIMA-TI-e-Election/internal/service"
)

func NewUserController(userService service.UserService) UserController {
	return &UserControllerImpl{
		UserService: userService,
	}
}

type UserControllerImpl struct {
	UserService service.UserService
}

func (controller *UserControllerImpl) Create(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Get request body and write it to userRequest
	userRequest := web.UserCreateRequest{}
	err := helper.ReadFromRequestBody(r, &userRequest)
	if err != nil {
		helper.WriteJSONDecodeError(w, err)
		return
	}

	// Call service
	userResponse, err := controller.UserService.Create(r.Context(), userRequest)
	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(err, "failed to create user")

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
	w.WriteHeader(http.StatusCreated)
	helper.WriteToResponseBody(w, web.WebSuccessResponse{
		Message: "User created successfully",
		Data:    userResponse,
	})
}

func (controller *UserControllerImpl) UpdateCurrent(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

	// Get request body and write it to userUpdateRequest
	userUpdateRequest := web.UserUpdateCurrentRequest{}
	err := helper.ReadFromRequestBody(r, &userUpdateRequest)
	if err != nil {
		helper.WriteJSONDecodeError(w, err)
		return
	}

	// Call service
	userResponse, err := controller.UserService.UpdateCurrent(r.Context(), cookie.SessionId, userUpdateRequest)
	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(err, "failed to update user")

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
		Message: "User updated successfully",
		Data:    userResponse,
	})
}

func (controller *UserControllerImpl) GetCurrent(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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
	userResponse, err := controller.UserService.GetCurrent(r.Context(), cookie.SessionId)
	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(err, "failed to get current user")

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

	// Write data and send the response
	w.WriteHeader(http.StatusOK)
	helper.WriteToResponseBody(w, web.WebSuccessResponse{
		Message: "Get current user successfully",
		Data:    userResponse,
	})
}

func (controller *UserControllerImpl) GetAll(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Call service
	users, err := controller.UserService.GetAll(r.Context())
	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(err, "failed to get all users")

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
		Message: "Success get all users",
		Data:    users,
	})
}

func (controller *UserControllerImpl) UpdateById(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Get user id from named parameter
	userId := params.ByName("userId")

	// Convert query params to int
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		appError.LogError(err, "failed to convert id to int")

		w.WriteHeader(http.StatusNotFound)
		helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
			"error": {
				Message: "User not found",
				Details: fmt.Sprintf("User with id '%v' does not exist", userId),
			},
		})
		return
	}

	// Get request body and write it to userRequest
	userRequest := web.UserUpdateByIdRequest{}
	err = helper.ReadFromRequestBody(r, &userRequest)
	if err != nil {
		helper.WriteJSONDecodeError(w, err)
		return
	}

	// Call service
	userResponse, err := controller.UserService.UpdateById(r.Context(), userIdInt, userRequest)
	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(err, "failed to update user by id")

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
		Message: "User updated successfully",
		Data:    userResponse,
	})
}

func (controller *UserControllerImpl) DeleteById(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Get user id from named parameter
	userId := params.ByName("userId")

	// Convert query params to int
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		appError.LogError(err, "failed to convert id to int")

		w.WriteHeader(http.StatusNotFound)
		helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
			"error": {
				Message: "User not found",
				Details: fmt.Sprintf("User with id '%v' does not exist", userId),
			},
		})
		return
	}

	// Call service
	err = controller.UserService.DeleteById(r.Context(), userIdInt)
	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(err, "failed to delete user by id")

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
		Message: "User deleted successfully",
		Data:    nil,
	})
}

func (controller *UserControllerImpl) BulkCreate(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Return 400 if request content-type isn't multipart/form-data
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		w.WriteHeader(http.StatusBadRequest)
		helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
			"error": {
				Message: "Bad request",
				Details: "Request content-type must be multipart/form-data",
			},
		})
		return
	}

	// Parse multipart form (max size 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		appError.LogError(err, "failed to parse multipart form")

		w.WriteHeader(http.StatusInternalServerError)
		helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
			"error": {
				Message: "Internal Server Error",
				Details: "Internal Server Error. Please try again later.",
			},
		})
		return
	}

	// Get the file with the field name "file"
	file, _, err := r.FormFile("file")
	if err != nil {
		appError.LogError(err, "file not found")

		w.WriteHeader(http.StatusNotFound)
		helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
			"error": {
				Message: "Not found",
				Details: "File not found in the request",
			},
		})
		return
	}
	defer file.Close()

	// Read the csv data
	var records []web.UserCreateRequest
	reader := csv.NewReader(file)

	// Set the Reader's FieldsPerRecord negative to skip the header
	reader.FieldsPerRecord = -1

	firstRow, err := reader.Read()
	if err != nil {
		appError.LogError(err, "file is not valid CSV format")

		w.WriteHeader(http.StatusBadRequest)
		helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
			"error": {
				Message: "Invalid CSV",
				Details: "Uploaded file is not a valid CSV (comma-separated values)",
			},
		})
		return
	}

	// Validate the first row
	if len(firstRow) != 4 {
		w.WriteHeader(http.StatusBadRequest)
		helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
			"error": {
				Message: "Invalid CSV structure in line 1",
				Details: "CSV must contain 4 columns: NIM, Full Name, Study Program, Phone Number",
			},
		})
		return
	}

	// Append the first row to the records
	records = append(records, web.UserCreateRequest{
		NIM:          firstRow[0],
		FullName:     firstRow[1],
		StudyProgram: firstRow[2],
		PhoneNumber:  firstRow[3],
	})

	i := 2
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			appError.LogError(err, "failed to read file")

			w.WriteHeader(http.StatusInternalServerError)
			helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
				"error": {
					Message: "Internal Server Error",
					Details: "Internal Server Error. Please try again later.",
				},
			})
			return
		}

		if len(record) != 4 {
			w.WriteHeader(http.StatusBadRequest)
			helper.WriteToResponseBody(w, map[string]web.WebFailedResponse{
				"error": {
					Message: fmt.Sprintf("Invalid CSV structure in line %d", i),
					Details: "CSV must contain 4 columns: NIM, Full Name, Study Program, Phone Number",
				},
			})
			return
		}

		records = append(records, web.UserCreateRequest{
			NIM:          record[0],
			FullName:     record[1],
			StudyProgram: record[2],
			PhoneNumber:  record[3],
		})

		i++
	}

	// Call service
	userResponses, err := controller.UserService.CreateBulk(r.Context(), records)
	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(err, "failed to create user")

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
	w.WriteHeader(http.StatusCreated)
	helper.WriteToResponseBody(w, web.WebSuccessResponse{
		Message: "All users created successfully",
		Data:    userResponses,
	})
}

func (controller *UserControllerImpl) GeneratePassword(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// Call service
	err := controller.UserService.GeneratePassword(r.Context())
	if err != nil {
		var customError *appError.AppError

		if errors.As(err, &customError) {
			appError.LogError(err, "failed to generate password and send to whatsapp")

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
	w.WriteHeader(http.StatusCreated)
	helper.WriteToResponseBody(w, web.WebSuccessResponse{
		Message: "Password generated successfully and has been sent to whatsapp",
		Data:    nil,
	})
}
