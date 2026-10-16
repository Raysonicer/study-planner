package api

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"studypilot/server/internal/service"
)

type errBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func fail(c echo.Context, status int, code, msg string) error {
	return c.JSON(status, errBody{Code: code, Message: msg})
}

// handleErr 统一错误映射：404/400/409/500。
func handleErr(c echo.Context, err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return fail(c, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, service.ErrInvalid):
		return fail(c, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, service.ErrConflict):
		return fail(c, http.StatusConflict, "conflict", err.Error())
	default:
		return fail(c, http.StatusInternalServerError, "internal", err.Error())
	}
}
