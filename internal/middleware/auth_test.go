package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

func signToken(t *testing.T, method jwt.SigningMethod, secret string, claims jwt.MapClaims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(method, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("cannot sign token: %v", err)
	}
	return token
}

func TestAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	future := time.Now().Add(time.Hour).Unix()
	past := time.Now().Add(-time.Hour).Unix()

	valid := signToken(t, jwt.SigningMethodHS256, testSecret,
		jwt.MapClaims{"staff_id": 1, "hospital_id": 5, "exp": future})
	wrongSecret := signToken(t, jwt.SigningMethodHS256, "other-secret",
		jwt.MapClaims{"staff_id": 1, "hospital_id": 5, "exp": future})
	expired := signToken(t, jwt.SigningMethodHS256, testSecret,
		jwt.MapClaims{"staff_id": 1, "hospital_id": 5, "exp": past})
	noHospital := signToken(t, jwt.SigningMethodHS256, testSecret,
		jwt.MapClaims{"staff_id": 1, "exp": future})
	wrongAlg := signToken(t, jwt.SigningMethodHS512, testSecret,
		jwt.MapClaims{"staff_id": 1, "hospital_id": 5, "exp": future})

	tests := []struct {
		name     string
		header   string
		want     int
		wantBody string
	}{
		{"valid token", "Bearer " + valid, http.StatusOK, `"hospital_id":5`},
		{"no header", "", http.StatusUnauthorized, ""},
		{"wrong prefix", "Token " + valid, http.StatusUnauthorized, ""},
		{"empty token", "Bearer ", http.StatusUnauthorized, ""},
		{"garbage token", "Bearer not-a-jwt", http.StatusUnauthorized, ""},
		{"signed with wrong secret", "Bearer " + wrongSecret, http.StatusUnauthorized, ""},
		{"expired token", "Bearer " + expired, http.StatusUnauthorized, ""},
		{"missing hospital_id claim", "Bearer " + noHospital, http.StatusUnauthorized, ""},
		{"unexpected signing algorithm", "Bearer " + wrongAlg, http.StatusUnauthorized, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/protected", Auth(testSecret), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"hospital_id": c.GetInt(HospitalIDKey)})
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.want {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.want, w.Body.String())
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("body %q does not contain %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}