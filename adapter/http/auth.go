package http

import (
	"crypto/subtle"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moycat/index/service"
)

func authMiddleware(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if err := validateBearerToken(header, token); err != nil {
			writeError(c, err)
			c.Abort()
			return
		}
		c.Next()
	}
}

func validateBearerToken(header, expected string) error {
	header = strings.TrimSpace(header)
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return nil
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return service.ErrUnauthorized
	}
	provided := strings.TrimSpace(parts[1])
	if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
		return service.ErrUnauthorized
	}
	return nil
}
