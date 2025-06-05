package errors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mhaatha/HIMA-TI-e-Election/config"
)

var (
	ErrForeignKeyViolation   = errors.New("foreign key violation")
	ErrUniqueFieldViolation  = errors.New("unique field violation")
	ErrEnumValidation        = errors.New("enum validation error")
	ErrTransaction           = errors.New("transaction error")
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidCredentials    = errors.New("invalid NIM or password")
	ErrSessionNotFound       = errors.New("session not found")
	ErrSessionExpired        = errors.New("session expired")
	ErrNIMAlreadyExists      = errors.New("NIM already exists")
	ErrNIMNotFound           = errors.New("NIM not found")
	ErrValidation            = errors.New("request validation error")
	ErrForbiddenAccess       = errors.New("forbidden access")
	ErrLoadDefaultConfig     = errors.New("failed to load s3 default config")
	ErrCreatePresignedPut    = errors.New("failed to create presigned for put object")
	ErrLoadEnvironmentConfig = errors.New("unable to load environment config")
	ErrNumberIsUsed          = errors.New("number is used")
	ErrPhotoKeyIsUsed        = errors.New("photo key is used")
	ErrInvalidPeriodRange    = errors.New("period must be a positive number and can't exceed 4 digits")
	ErrInvalidPeriodSyntax   = errors.New("period can't contain characters")
	ErrCandidateHasVotes     = errors.New("candidate has votes")
)

type AppError struct {
	StatusCode int    // HTTP Status Code
	Message    string // Message e.g: "Invalid credentials"
	Details    string // Details e.g: "The NIM or password is incorrect"
	Err        error  // Underlying error e.g: user with NIM 23666101 not found: no rows in result set
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("(%v) %v - %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("%v: %v", e.Message, e.Details)
}

func NewAppError(code int, message, details string, err error) *AppError {
	return &AppError{
		StatusCode: code,
		Message:    message,
		Details:    details,
		Err:        err,
	}
}

// LogError logs the error with the given message.
// If the error is nil, this function does nothing.
func LogError(err error, message string) {
	if err != nil {
		config.Log.WithError(err).Error(message)
	}
}

// formatValidationDetails generates a detailed error message string from
// a slice of ValidationErrors. Each error in the slice is formatted to
// include the field name, the invalid value, and the validation rule
// that was violated. The resulting string contains all errors separated
// by semicolons, with the trailing semicolon removed.
func FormatValidationDetails(ve validator.ValidationErrors) string {
	var details strings.Builder
	for _, fe := range ve {
		switch fe.Tag() {
		case "min":
			details.WriteString(fmt.Sprintf(
				"[%s]: '%v' failed validation 'minLength' (minimum length: %s); ",
				fe.Field(), fe.Value(), fe.Param(),
			))
		case "max":
			details.WriteString(fmt.Sprintf(
				"[%s]: '%v' failed validation 'maxLength' (maximum length: %s); ",
				fe.Field(), fe.Value(), fe.Param(),
			))
		case "required":
			details.WriteString(fmt.Sprintf(
				"[%s]: '%v' failed validation 'required'; ",
				fe.Field(), fe.Value(),
			))
		default:
			details.WriteString(fmt.Sprintf(
				"[%s]: '%v' failed validation '%s'; ",
				fe.Field(), fe.Value(), fe.Tag(),
			))
		}
	}
	return strings.TrimSuffix(details.String(), "; ")
}

func FormatValidationDetailsWithRow(ve validator.ValidationErrors, row int) string {
	var details strings.Builder
	details.WriteString(fmt.Sprintf("CSV's line %d: ", row))

	for _, fe := range ve {
		switch fe.Tag() {
		case "min":
			details.WriteString(fmt.Sprintf(
				"[%s]: '%v' failed validation 'minLength' (minimum length: %s); ",
				fe.Field(), fe.Value(), fe.Param(),
			))
		case "max":
			details.WriteString(fmt.Sprintf(
				"[%s]: '%v' failed validation 'maxLength' (maximum length: %s); ",
				fe.Field(), fe.Value(), fe.Param(),
			))
		case "required":
			details.WriteString(fmt.Sprintf(
				"[%s]: '%v' failed validation 'required'; ",
				fe.Field(), fe.Value(),
			))
		default:
			details.WriteString(fmt.Sprintf(
				"[%s]: '%v' failed validation '%s'; ",
				fe.Field(), fe.Value(), fe.Tag(),
			))
		}
	}
	return strings.TrimSuffix(details.String(), "; ")
}
