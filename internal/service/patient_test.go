package service

import (
	"context"
	"errors"
	"testing"

	"github.com/JONE500I007/agnos-hospital-api/internal/model"
)

type fakePatientStore struct {
	patients []model.Patient
	patient  *model.Patient
	err      error

	gotHospitalID int
	gotFilter     model.PatientSearch
	gotID         string
}

func (f *fakePatientStore) Search(ctx context.Context, hospitalID int, filter model.PatientSearch) ([]model.Patient, error) {
	f.gotHospitalID = hospitalID
	f.gotFilter = filter
	return f.patients, f.err
}

func (f *fakePatientStore) FindByID(ctx context.Context, hospitalID int, id string) (*model.Patient, error) {
	f.gotHospitalID = hospitalID
	f.gotID = id
	return f.patient, f.err
}

func TestPatientServiceSearch(t *testing.T) {
	t.Run("passes hospital and filter to the store", func(t *testing.T) {
		store := &fakePatientStore{patients: []model.Patient{{NationalID: "1"}}}
		svc := NewPatientService(store)

		got, err := svc.Search(context.Background(), 4, model.PatientSearch{NationalID: "1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].NationalID != "1" {
			t.Errorf("patients = %+v", got)
		}
		if store.gotHospitalID != 4 || store.gotFilter.NationalID != "1" {
			t.Errorf("store got hospital=%d filter=%+v", store.gotHospitalID, store.gotFilter)
		}
	})

	t.Run("passes store errors through", func(t *testing.T) {
		dbErr := errors.New("db down")
		svc := NewPatientService(&fakePatientStore{err: dbErr})
		if _, err := svc.Search(context.Background(), 1, model.PatientSearch{}); !errors.Is(err, dbErr) {
			t.Errorf("error = %v, want %v", err, dbErr)
		}
	})
}

func TestPatientServiceFindByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		store := &fakePatientStore{patient: &model.Patient{NationalID: "1"}}
		svc := NewPatientService(store)

		got, err := svc.FindByID(context.Background(), 4, "1")
		if err != nil || got == nil || got.NationalID != "1" {
			t.Fatalf("got %+v, err %v", got, err)
		}
		if store.gotHospitalID != 4 || store.gotID != "1" {
			t.Errorf("store got hospital=%d id=%q", store.gotHospitalID, store.gotID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc := NewPatientService(&fakePatientStore{err: model.ErrPatientNotFound})
		if _, err := svc.FindByID(context.Background(), 4, "x"); !errors.Is(err, model.ErrPatientNotFound) {
			t.Errorf("error = %v, want ErrPatientNotFound", err)
		}
	})
}