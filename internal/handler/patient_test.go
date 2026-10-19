package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/JONE500I007/agnos-hospital-api/internal/middleware"
	"github.com/JONE500I007/agnos-hospital-api/internal/model"
	"github.com/gin-gonic/gin"
)

const testHospitalID = 7

type fakePatientService struct {
	patients []model.Patient
	patient  *model.Patient
	err      error

	gotHospitalID int
	gotFilter     model.PatientSearch
	gotID         string
}

func (f *fakePatientService) Search(ctx context.Context, hospitalID int, filter model.PatientSearch) ([]model.Patient, error) {
	f.gotHospitalID = hospitalID
	f.gotFilter = filter
	return f.patients, f.err
}

func (f *fakePatientService) FindByID(ctx context.Context, hospitalID int, id string) (*model.Patient, error) {
	f.gotHospitalID = hospitalID
	f.gotID = id
	return f.patient, f.err
}

func newPatientRouter(svc PatientService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.HospitalIDKey, testHospitalID)
		c.Next()
	})
	h := NewPatientHandler(svc)
	r.POST("/patient/search", h.Search)
	r.GET("/patient/search/:id", h.GetByID)
	return r
}

func TestPatientSearch(t *testing.T) {
	tests := []struct {
		name string
		body string
		err  error
		want int
	}{
		{"success with filter", `{"national_id":"1100100000011"}`, nil, http.StatusOK},
		{"empty object", `{}`, nil, http.StatusOK},
		{"empty body", ``, nil, http.StatusOK},
		{"invalid json", `{`, nil, http.StatusBadRequest},
		{"invalid date format", `{"date_of_birth":"12-05-1990"}`, nil, http.StatusBadRequest},
		{"valid date", `{"date_of_birth":"1990-05-12"}`, nil, http.StatusOK},
		{"internal error", `{}`, errors.New("db down"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newPatientRouter(&fakePatientService{err: tt.err})
			w := doJSON(r, http.MethodPost, "/patient/search", tt.body)
			if w.Code != tt.want {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.want, w.Body.String())
			}
		})
	}
}

func TestPatientSearchUsesHospitalFromToken(t *testing.T) {
	svc := &fakePatientService{}
	r := newPatientRouter(svc)

	doJSON(r, http.MethodPost, "/patient/search", `{"national_id":"1100100000011","hospital_id":99}`)

	if svc.gotHospitalID != testHospitalID {
		t.Errorf("hospitalID = %d, want %d", svc.gotHospitalID, testHospitalID)
	}
	if svc.gotFilter.NationalID != "1100100000011" {
		t.Errorf("national_id = %q, want %q", svc.gotFilter.NationalID, "1100100000011")
	}
}

func TestPatientGetByID(t *testing.T) {
	tests := []struct {
		name string
		svc  *fakePatientService
		want int
	}{
		{"found", &fakePatientService{patient: &model.Patient{NationalID: "1100100000011"}}, http.StatusOK},
		{"not found", &fakePatientService{err: model.ErrPatientNotFound}, http.StatusNotFound},
		{"internal error", &fakePatientService{err: errors.New("db down")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newPatientRouter(tt.svc)
			w := doJSON(r, http.MethodGet, "/patient/search/1100100000011", "")
			if w.Code != tt.want {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.want, w.Body.String())
			}
			if tt.svc.gotID != "1100100000011" || tt.svc.gotHospitalID != testHospitalID {
				t.Errorf("service called with id=%q hospital=%d", tt.svc.gotID, tt.svc.gotHospitalID)
			}
		})
	}
}