package api

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nwasiq/fieldops/backend/internal/services"
)

type successBody struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

type errorBody struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func respond(c *gin.Context, status int, data any) {
	c.JSON(status, successBody{Success: true, Data: data})
}

func respondError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, errorBody{Success: false, Error: message})
}

// writeInternalError is the one path for unexpected failures: the error is
// logged with the request path and the client gets a generic 500.
func writeInternalError(c *gin.Context, logger *slog.Logger, err error) {
	logger.Error("request failed",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"error", err.Error(),
	)
	respondError(c, http.StatusInternalServerError, "internal server error")
}

// fail maps a service error to its status code, or to a logged 500.
func fail(c *gin.Context, logger *slog.Logger, err error) {
	switch services.KindOf(err) {
	case services.KindInvalid:
		respondError(c, http.StatusBadRequest, err.Error())
	case services.KindUnauthorized:
		respondError(c, http.StatusUnauthorized, err.Error())
	case services.KindForbidden:
		respondError(c, http.StatusForbidden, err.Error())
	case services.KindNotFound:
		respondError(c, http.StatusNotFound, err.Error())
	case services.KindConflict:
		respondError(c, http.StatusConflict, err.Error())
	default:
		writeInternalError(c, logger, err)
	}
}

var errBadID = errors.New("id must be a positive integer")

func pathID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, errBadID
	}
	return uint(id), nil
}

func queryInt(c *gin.Context, name string) int {
	value, _ := strconv.Atoi(c.Query(name))
	return value
}

func pageFromQuery(c *gin.Context) services.Page {
	return services.NormalisePage(queryInt(c, "page"), queryInt(c, "page_size"))
}
