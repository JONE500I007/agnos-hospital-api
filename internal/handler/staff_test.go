package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JONE500I007/agnos-hospital-api/internal/model"
	"github.com/gin-gonic/gin"
)

type fakeStaffService struct {
	createErr error
	token     string
	loginErr  error
}

func (f *fakeStaffService) Create(ctx context.Context, username, password, hospital string) error {
	return f.createErr
}

func (f *fakeStaffService) Login(ctx context.Context, username, password string) (string, error) {
	return f.token, f.loginErr
}

func newStaffRouter(svc StaffService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewStaffHandler(svc)
	r.POST("/staff/create", h.Create)
	r.POST("/staff/login", h.Login)
	return r
}

func doJSON(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestStaffCreate(t *testing.T) {
	const valid = `{"username":"nurse1","password":"pass1234","hospital":"Hospital A"}`

	tests := []struct {
		name string
		body string
		err  error
		want int
	}{
		{"success", valid, nil, http.StatusCreated},
		{"missing password", `{"username":"nurse1","hospital":"Hospital A"}`, nil, http.StatusBadRequest},
		{"password too short", `{"username":"nurse1","password":"123","hospital":"Hospital A"}`, nil, http.StatusBadRequest},
		{"missing hospital", `{"username":"nurse1","password":"pass1234"}`, nil, http.StatusBadRequest},
		{"invalid json", `{`, nil, http.StatusBadRequest},
		{"duplicate username", valid, model.ErrUsernameTaken, http.StatusConflict},
		{"unknown hospital", valid, model.ErrHospitalNotFound, http.StatusBadRequest},
		{"internal error", valid, errors.New("db down"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newStaffRouter(&fakeStaffService{createErr: tt.err})
			w := doJSON(r, http.MethodPost, "/staff/create", tt.body)
			if w.Code != tt.want {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.want, w.Body.String())
			}
		})
	}
}

func TestStaffLogin(t *testing.T) {
	const valid = `{"username":"nurse1","password":"pass1234"}`

	tests := []struct {
		name     string
		body     string
		token    string
		err      error
		want     int
		wantBody string
	}{
		{"success", valid, "tok123", nil, http.StatusOK, `"token":"tok123"`},
		{"wrong credentials", valid, "", model.ErrInvalidCredentials, http.StatusUnauthorized, ""},
		{"missing username", `{"password":"pass1234"}`, "", nil, http.StatusBadRequest, ""},
		{"missing password", `{"username":"nurse1"}`, "", nil, http.StatusBadRequest, ""},
		{"invalid json", `{`, "", nil, http.StatusBadRequest, ""},
		{"internal error", valid, "", errors.New("db down"), http.StatusInternalServerError, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newStaffRouter(&fakeStaffService{token: tt.token, loginErr: tt.err})
			w := doJSON(r, http.MethodPost, "/staff/login", tt.body)
			if w.Code != tt.want {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.want, w.Body.String())
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("body %q does not contain %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}