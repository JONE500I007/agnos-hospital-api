package service

import (
	"context"
	"errors"

	"github.com/JONE500I007/agnos-hospital-api/internal/model"
)

type PatientStore interface {
	Search(ctx context.Context, hospitalID int, f model.PatientSearch) ([]model.Patient, error)
	FindByID(ctx context.Context, hospitalID int, id string) (*model.Patient, error)
	HospitalName(ctx context.Context, hospitalID int) (string, error)
	Save(ctx context.Context, hospitalID int, p model.Patient) error
}

type HISClient interface {
	FindPatient(ctx context.Context, id string) (*model.Patient, error)
}

type PatientService struct {
	store PatientStore
	his   map[string]HISClient // keyed by hospital name
}

func NewPatientService(store PatientStore, his map[string]HISClient) *PatientService {
	return &PatientService{store: store, his: his}
}

func (s *PatientService) Search(ctx context.Context, hospitalID int, f model.PatientSearch) ([]model.Patient, error) {
	return s.store.Search(ctx, hospitalID, f)
}

func (s *PatientService) FindByID(ctx context.Context, hospitalID int, id string) (*model.Patient, error) {
	p, err := s.store.FindByID(ctx, hospitalID, id)
	if !errors.Is(err, model.ErrPatientNotFound) {
		return p, err
	}

	name, err := s.store.HospitalName(ctx, hospitalID)
	if err != nil {
		return nil, err
	}
	client, ok := s.his[name]
	if !ok {
		return nil, model.ErrPatientNotFound
	}

	p, err = client.FindPatient(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.store.Save(ctx, hospitalID, *p); err != nil {
		return nil, err
	}
	return p, nil
}
