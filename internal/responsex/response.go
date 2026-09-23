package responsex

import (
	"context"
	"net/http"

	"fishing-notes-admin-api/internal/errorsx"
)

type Envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func OkHandler(_ context.Context, v any) any {
	return Envelope{
		Code:    0,
		Message: "ok",
		Data:    v,
	}
}

func ErrorHandler(_ context.Context, err error) (int, any) {
	if apiErr, ok := errorsx.As(err); ok {
		return apiErr.StatusCode, Envelope{
			Code:    apiErr.Code,
			Message: apiErr.Message,
		}
	}

	return http.StatusInternalServerError, Envelope{
		Code:    errorsx.CodeInternal,
		Message: "internal server error",
	}
}
