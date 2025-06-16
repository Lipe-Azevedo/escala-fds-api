package auth

import (
	"escala-fds-api/internal/constants"
	"escala-fds-api/pkg/ierr"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			err := ierr.NewUnauthorizedError("authorization header is required")
			c.AbortWithStatusJSON(err.Code, err)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			err := ierr.NewUnauthorizedError("authorization header format must be Bearer {token}")
			c.AbortWithStatusJSON(err.Code, err)
			return
		}

		tokenString := parts[1]
		claims := &jwt.MapClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})

		if err != nil {
			errRest := ierr.NewUnauthorizedError("invalid or expired token")
			c.AbortWithStatusJSON(errRest.Code, errRest)
			return
		}

		if !token.Valid {
			errRest := ierr.NewUnauthorizedError("invalid token")
			c.AbortWithStatusJSON(errRest.Code, errRest)
			return
		}

		idFloat, ok := (*claims)["id"].(float64)
		if !ok {
			errRest := ierr.NewUnauthorizedError("invalid user id in token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, errRest)
			return
		}

		userType, ok := (*claims)["user_type"].(string)
		if !ok {
			errRest := ierr.NewUnauthorizedError("invalid user type in token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, errRest)
			return
		}

		team, _ := (*claims)["team"].(string)

		c.Set(constants.JwtUserIdKey, uint(idFloat))
		c.Set(constants.JwtUserTypeKey, userType)
		c.Set(constants.JwtTeamKey, team)
		c.Next()
	}
}

func GetUserIDFromContext(c *gin.Context) (uint, *ierr.RestErr) {
	id, ok := c.Get(constants.JwtUserIdKey)
	if !ok {
		return 0, ierr.NewInternalServerError("user id not found in context")
	}
	userId, ok := id.(uint)
	if !ok {
		return 0, ierr.NewInternalServerError("invalid user id type in context")
	}
	return userId, nil
}

func GetUserTypeFromContext(c *gin.Context) (string, *ierr.RestErr) {
	userType, ok := c.Get(constants.JwtUserTypeKey)
	if !ok {
		return "", ierr.NewInternalServerError("user type not found in context")
	}
	userTypeStr, ok := userType.(string)
	if !ok {
		return "", ierr.NewInternalServerError("invalid user type in context")
	}
	return userTypeStr, nil
}

func GetUserTeamFromContext(c *gin.Context) (string, *ierr.RestErr) {
	team, ok := c.Get(constants.JwtTeamKey)
	if !ok {
		return "", ierr.NewInternalServerError("user team not found in context")
	}
	teamStr, ok := team.(string)
	if !ok {
		return "", ierr.NewInternalServerError("invalid user team type in context")
	}
	return teamStr, nil
}
