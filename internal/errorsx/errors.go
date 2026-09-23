package errorsx

import "net/http"

const (
	CodeBadRequest   = 10000
	CodeUnauthorized = 10001
	CodeForbidden    = 10003
	CodeNotFound     = 10004
	CodeInternal     = 10005
)

type Error struct {
	Code       int
	Message    string
	StatusCode int
}

func (e *Error) Error() string {
	return e.Message
}

func As(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}

	apiErr, ok := err.(*Error)
	return apiErr, ok
}

func BadRequest(message string) error {
	return &Error{Code: CodeBadRequest, Message: message, StatusCode: http.StatusBadRequest}
}

func Unauthorized(message string) error {
	return &Error{Code: CodeUnauthorized, Message: message, StatusCode: http.StatusUnauthorized}
}

func Forbidden(message string) error {
	return &Error{Code: CodeForbidden, Message: message, StatusCode: http.StatusForbidden}
}

func NotFound(message string) error {
	return &Error{Code: CodeNotFound, Message: message, StatusCode: http.StatusNotFound}
}

func Internal(message string) error {
	return &Error{Code: CodeInternal, Message: message, StatusCode: http.StatusInternalServerError}
}
