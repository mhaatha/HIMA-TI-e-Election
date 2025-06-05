package helper

import (
	"encoding/json"
	"fmt"
	"net/http"

	myErrors "github.com/mhaatha/HIMA-TI-e-Election/errors"
	"github.com/mhaatha/HIMA-TI-e-Election/model/web"
)

// ReadFromRequestBody reads the request body and stores it in the result parameter
func ReadFromRequestBody(r *http.Request, result interface{}) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(result)
	if err != nil {
		myErrors.LogError(err, "error when reading request body")
		return err
	}

	return nil
}

// WriteToResponseBody writes the result parameter to the response body
func WriteToResponseBody(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(response); err != nil {
		myErrors.LogError(err, "error when writing response body")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// WriteJSONDecodeError writes an error response to the client if json decoding fails.
func WriteJSONDecodeError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *json.UnmarshalTypeError:
		w.WriteHeader(http.StatusBadRequest)
		WriteToResponseBody(w, map[string]web.WebFailedResponse{
			"error": {
				Message: "Invalid request body",
				Details: fmt.Sprintf("Invalid type %v for field %v", e.Value, e.Field),
			},
		})
	case *json.SyntaxError:
		w.WriteHeader(http.StatusBadRequest)
		WriteToResponseBody(w, map[string]web.WebFailedResponse{
			"error": {
				Message: "Invalid request body",
				Details: e.Error(),
			},
		})
	default:
		w.WriteHeader(http.StatusInternalServerError)
		WriteToResponseBody(w, map[string]web.WebFailedResponse{
			"error": {
				Message: "Internal Server Error",
				Details: "Internal Server Error. Please try again later.",
			},
		})
	}
}
