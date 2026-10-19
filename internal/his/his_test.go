package his

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JONE500I007/agnos-hospital-api/internal/model"
)

func newTestHIS(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/patient/search/1100100000044":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"first_name_en":"Anan","national_id":"1100100000044","gender":"M"}`))
		case "/patient/search/bad-json":
			w.Write([]byte(`{`))
		case "/patient/search/server-error":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestHTTPClientFindPatient(t *testing.T) {
	srv := newTestHIS(t)
	client := NewHTTPClient(srv.URL+"/", time.Second)

	t.Run("found", func(t *testing.T) {
		p, err := client.FindPatient(context.Background(), "1100100000044")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.FirstNameEN != "Anan" || p.NationalID != "1100100000044" || p.Gender != "M" {
			t.Errorf("patient = %+v", p)
		}
	})

	tests := []struct {
		name string
		id   string
		want error
	}{
		{"not found", "9999999999999", model.ErrPatientNotFound},
		{"server error", "server-error", model.ErrHISUnavailable},
		{"invalid json", "bad-json", model.ErrHISUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := client.FindPatient(context.Background(), tt.id)
			if !errors.Is(err, tt.want) {
				t.Errorf("error = %v, want %v", err, tt.want)
			}
			if p != nil {
				t.Errorf("patient = %+v, want nil", p)
			}
		})
	}
}

func TestHTTPClientUnreachable(t *testing.T) {
	srv := newTestHIS(t)
	url := srv.URL
	srv.Close()

	_, err := NewHTTPClient(url, time.Second).FindPatient(context.Background(), "1100100000044")
	if !errors.Is(err, model.ErrHISUnavailable) {
		t.Errorf("error = %v, want ErrHISUnavailable", err)
	}
}

func TestMockClient(t *testing.T) {
	client := NewMockClient(SampleHospitalAPatients()...)

	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{"by national id", "1100100000044", nil},
		{"by passport id", "AA1234567", nil},
		{"unknown id", "0000000000000", model.ErrPatientNotFound},
		{"empty id", "", model.ErrPatientNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := client.FindPatient(context.Background(), tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && p.NationalID != "1100100000044" {
				t.Errorf("patient = %+v", p)
			}
		})
	}
}
