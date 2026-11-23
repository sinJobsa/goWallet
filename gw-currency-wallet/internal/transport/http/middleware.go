package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sinJobsa/goWallet/gw-currency-wallet/internal/auth"
)

const userIDKey = "user_id"


func JWTMiddleware(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "authorization header required"})
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid authorization header"})
			return
		}


		claims, err := authService.Parse(parts[1])
		if errors.Is(err, auth.ErrExpiredToken) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "token expired"})
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
			return
		}

		c.Set(userIDKey, claims.UserID)
		c.Next()
	}
}
