package apierrors

import "net/http"

// APIError represents an error that is returned to the client.
type APIError struct {
	Code       string `json:"code"`    // error code that can be used to identify the error
	Message    string `json:"message"` // detailed description of the error
	HTTPStatus int    `json:"-"`       // http status code. It is not included in the response body
}

// NewAPIError creates a new APIError.
func NewAPIError(code string, message string, httpStatus int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Error returns the error message.
func (e *APIError) Error() string {
	return e.Message
}

var (

	// ErrInternalServer is returned when an internal server error occurs.
	ErrInternalServer = NewAPIError("INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)

	// ErrInvalidBody is returned when the request body is invalid.
	ErrInvalidBody = NewAPIError("INVALID_BODY", "invalid request body", http.StatusBadRequest)

	// ErrUserNotFound is returned when a user is not found.
	ErrUserNotFound = NewAPIError("USER_NOT_FOUND", "user not found", http.StatusBadRequest)

	// ErrUserAlreadyExists is returned when a user already exists.
	ErrUserAlreadyExists = NewAPIError("USER_ALREADY_EXISTS", "user already exists", http.StatusBadRequest)

	// ErrInvalidCredentials is returned when the credentials are invalid.
	ErrInvalidCredentials = NewAPIError("INVALID_CREDENTIALS", "invalid credentials", http.StatusUnauthorized)

	// ErrCompanyNotFound is returned when a company is not found.
	ErrCompanyNotFound = NewAPIError("COMPANY_NOT_FOUND", "company not found", http.StatusBadRequest)

	// ErrInvalidUUID is returned when the UUID is invalid.
	ErrInvalidUUID = NewAPIError("INVALID_UUID", "invalid UUID", http.StatusBadRequest)

	// ErrTokenNotFound is returned when a token is not found.
	ErrTokenNotFound = NewAPIError("TOKEN_NOT_FOUND", "token not found", http.StatusBadRequest)

	// ErrInvalidToken is returned when the token is invalid.
	ErrInvalidToken = NewAPIError("INVALID_TOKEN", "invalid token", http.StatusUnauthorized)

	// ErrUnauthorized is returned when the user is not authorized.
	ErrUnauthorized = NewAPIError("UNAUTHORIZED", "unauthorized", http.StatusUnauthorized)

	// ErrTokenExpired is returned when the token is expired.
	ErrTokenExpired = NewAPIError("TOKEN_EXPIRED", "token expired", http.StatusUnauthorized)

	// ErrCompanyIDRequired is returned when the company ID is required.
	ErrCompanyIDRequired = NewAPIError("COMPANY_ID_REQUIRED", "company ID is required", http.StatusBadRequest)

	// ErrCreatingEvent is returned when an error occurs while creating an event.
	ErrCreatingEvent = NewAPIError("CREATING_EVENT", "error creating event", http.StatusInternalServerError)
)
