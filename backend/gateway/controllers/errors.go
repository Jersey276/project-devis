package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// serviceErrors holds the upstream-code → HTTP-response mapping for one
// backend service. Each controller defines its own instance; the methods
// here produce a uniform JSON shape across services.
type serviceErrors struct {
	codes              map[int32]codeMapping
	unavailableMessage string
}

type codeMapping struct {
	Status  int
	Message string
}

// FieldError is the JSON shape for a single field-level validation error.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// reply emits the mapped status + message for `code`, or a generic 500 if
// the code is unknown.
func (s *serviceErrors) reply(c *gin.Context, code int32) {
	if mapped, ok := s.codes[code]; ok {
		c.JSON(mapped.Status, gin.H{"success": false, "message": mapped.Message, "code": code})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Une erreur inconnue est survenue.", "code": code})
}

// replyWithValidation emits the mapped status + message for `code` and
// includes a field-level `errors` array when errors is non-empty.
func (s *serviceErrors) replyWithValidation(c *gin.Context, code int32, errors []FieldError) {
	if mapped, ok := s.codes[code]; ok {
		body := gin.H{"success": false, "message": mapped.Message, "code": code}
		if len(errors) > 0 {
			body["errors"] = errors
		}
		c.JSON(mapped.Status, body)
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Une erreur inconnue est survenue.", "code": code})
}

// unavailable emits a 502 for transport-level failures (gRPC dial/timeout).
func (s *serviceErrors) unavailable(c *gin.Context) {
	c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": s.unavailableMessage})
}
