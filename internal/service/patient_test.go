package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/JONE500I007/agnos-hospital-api/internal/model"
)

type fakePatientStore struct {
	patients     []model.Patient
	patient      *model.Patient
	err          error
	hospitalName string
	nameErr      error
	saveErr      error

	gotHospitalID   int
	gotFilter       model.PatientSearch
	gotID           string
	saved           []model.Patient
	savedHospitalID int
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

func (f *fakePatientStore) HospitalName(ctx context.Context, hospitalID int) (string, error) {
	return f.hospitalName, f.nameErr
}

func (f *fakePatientStore) Save(ctx context.Context, hospitalID int, p model.Patient) error {
	f.savedHospitalID = hospitalID
	f.saved = append(f.saved, p)
	return f.saveErr
}

type fakeHIS struct {
	patient *model.Patient
	err     error

	calls int
	gotID string
}

func (f *fakeHIS) FindPatient(ctx context.Context, id string) (*model.Patient, error) {
	f.calls++
	f.gotID = id
	return f.patient, f.err
}

func TestPatientServiceSearch(t *testing.T) {
	t.Run("passes hospital and filter to the store", func(t *testing.T) {
		store := &fakePatientStore{patients: []model.Patient{{NationalID: "1"}}}
		svc := NewPatientService(store, nil)

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
		svc := NewPatientService(&fakePatientStore{err: dbErr}, nil)
		if _, err := svc.Search(context.Background(), 1, model.PatientSearch{}); !errors.Is(err, dbErr) {
			t.Errorf("error = %v, want %v", err, dbErr)
		}
	})
}

func TestPatientServiceFindByID(t *testing.T) {
	hisPatient := &model.Patient{NationalID: "1100100000044", FirstNameEN: "Anan"}

	t.Run("found in the database: HIS is not called", func(t *testing.T) {
		store := &fakePatientStore{patient: &model.Patient{NationalID: "1"}}
		his := &fakeHIS{patient: hisPatient}
		svc := NewPatientService(store, map[string]HISClient{"Hospital A": his})

		got, err := svc.FindByID(context.Background(), 4, "1")
		if err != nil || got == nil || got.NationalID != "1" {
			t.Fatalf("got %+v, err %v", got, err)
		}
		if store.gotHospitalID != 4 || store.gotID != "1" {
			t.Errorf("store got hospital=%d id=%q", store.gotHospitalID, store.gotID)
		}
		if his.calls != 0 {
			t.Errorf("HIS called %d times, want 0", his.calls)
		}
	})

	t.Run("not in the database: fetched from the hospital's HIS and saved", func(t *testing.T) {
		store := &fakePatientStore{err: model.ErrPatientNotFound, hospitalName: "Hospital A"}
		his := &fakeHIS{patient: hisPatient}
		svc := NewPatientService(store, map[string]HISClient{"Hospital A": his})

		got, err := svc.FindByID(context.Background(), 4, "1100100000044")
		if err != nil || got == nil || got.FirstNameEN != "Anan" {
			t.Fatalf("got %+v, err %v", got, err)
		}
		if his.gotID != "1100100000044" {
			t.Errorf("HIS got id %q", his.gotID)
		}
		if len(store.saved) != 1 || store.savedHospitalID != 4 {
			t.Errorf("saved %d patients for hospital %d, want 1 for hospital 4", len(store.saved), store.savedHospitalID)
		}
	})

	t.Run("hospital without a HIS: not found and no HIS call", func(t *testing.T) {
		store := &fakePatientStore{err: model.ErrPatientNotFound, hospitalName: "Hospital B"}
		his := &fakeHIS{patient: hisPatient}
		svc := NewPatientService(store, map[string]HISClient{"Hospital A": his})

		if _, err := svc.FindByID(context.Background(), 5, "1100100000044"); !errors.Is(err, model.ErrPatientNotFound) {
			t.Errorf("error = %v, want ErrPatientNotFound", err)
		}
		if his.calls != 0 {
			t.Errorf("HIS called %d times, want 0", his.calls)
		}
	})

	t.Run("not in the database nor in the HIS", func(t *testing.T) {
		store := &fakePatientStore{err: model.ErrPatientNotFound, hospitalName: "Hospital A"}
		svc := NewPatientService(store, map[string]HISClient{"Hospital A": &fakeHIS{err: model.ErrPatientNotFound}})

		if _, err := svc.FindByID(context.Background(), 4, "x"); !errors.Is(err, model.ErrPatientNotFound) {
			t.Errorf("error = %v, want ErrPatientNotFound", err)
		}
		if len(store.saved) != 0 {
			t.Error("nothing should be saved")
		}
	})

	t.Run("HIS unavailable", func(t *testing.T) {
		store := &fakePatientStore{err: model.ErrPatientNotFound, hospitalName: "Hospital A"}
		hisErr := fmt.Errorf("%w: timeout", model.ErrHISUnavailable)
		svc := NewPatientService(store, map[string]HISClient{"Hospital A": &fakeHIS{err: hisErr}})

		if _, err := svc.FindByID(context.Background(), 4, "x"); !errors.Is(err, model.ErrHISUnavailable) {
			t.Errorf("error = %v, want ErrHISUnavailable", err)
		}
		if len(store.saved) != 0 {
			t.Error("nothing should be saved")
		}
	})

	t.Run("database error is returned without calling the HIS", func(t *testing.T) {
		dbErr := errors.New("db down")
		his := &fakeHIS{patient: hisPatient}
		svc := NewPatientService(&fakePatientStore{err: dbErr}, map[string]HISClient{"Hospital A": his})

		if _, err := svc.FindByID(context.Background(), 4, "x"); !errors.Is(err, dbErr) {
			t.Errorf("error = %v, want %v", err, dbErr)
		}
		if his.calls != 0 {
			t.Errorf("HIS called %d times, want 0", his.calls)
		}
	})

	t.Run("hospital lookup fails", func(t *testing.T) {
		dbErr := errors.New("db down")
		store := &fakePatientStore{err: model.ErrPatientNotFound, nameErr: dbErr}
		svc := NewPatientService(store, map[string]HISClient{"Hospital A": &fakeHIS{patient: hisPatient}})

		if _, err := svc.FindByID(context.Background(), 4, "x"); !errors.Is(err, dbErr) {
			t.Errorf("error = %v, want %v", err, dbErr)
		}
	})

	t.Run("saving the HIS result fails", func(t *testing.T) {
		dbErr := errors.New("insert failed")
		store := &fakePatientStore{err: model.ErrPatientNotFound, hospitalName: "Hospital A", saveErr: dbErr}
		svc := NewPatientService(store, map[string]HISClient{"Hospital A": &fakeHIS{patient: hisPatient}})

		if _, err := svc.FindByID(context.Background(), 4, "1100100000044"); !errors.Is(err, dbErr) {
			t.Errorf("error = %v, want %v", err, dbErr)
		}
	})
}
