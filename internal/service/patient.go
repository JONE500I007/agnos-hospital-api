package service

import (
	"context"

	"github.com/JONE500I007/agnos-hospital-api/internal/model"
)

type PatientStore interface {
	Search(ctx context.Context, hospitalID int, f model.PatientSearch) ([]model.Patient, error)
	FindByID(ctx context.Context, hospitalID int, id string) (*model.Patient, error)
}

type PatientService struct {
	store PatientStore
}

func NewPatientService(store PatientStore) *PatientService {
	return &PatientService{store: store}
}



func (s *PatientService) Search(ctx context.Context, hospitalID int, f model.PatientSearch) ([]model.Patient, error) {
	return s.store.Search(ctx, hospitalID, f)
}

func (s *PatientService) FindByID(ctx context.Context, hospitalID int, id string) (*model.Patient, error) {
	return s.store.FindByID(ctx, hospitalID, id)
}