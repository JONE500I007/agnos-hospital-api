package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const HospitalIDKey = "hospital_id"

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !ok || tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims,
			func(t *jwt.Token) (any, error) { return []byte(secret), nil },
			jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}

		hospitalID, ok := claims["hospital_id"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}

		c.Set(HospitalIDKey, int(hospitalID))
		c.Next()
	}
}