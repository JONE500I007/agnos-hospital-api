package his

import (
	"context"

	"github.com/JONE500I007/agnos-hospital-api/internal/model"
)

type MockClient struct {
	patients []model.Patient
}

func NewMockClient(patients ...model.Patient) *MockClient {
	return &MockClient{patients: patients}
}

func (m *MockClient) FindPatient(ctx context.Context, id string) (*model.Patient, error) {
	if id == "" {
		return nil, model.ErrPatientNotFound
	}
	for _, p := range m.patients {
		if p.NationalID == id || p.PassportID == id {
			return &p, nil
		}
	}
	return nil, model.ErrPatientNotFound
}

func SampleHospitalAPatients() []model.Patient {
	return []model.Patient{
		{
			FirstNameTH: "อนันต์",
			LastNameTH:  "สุขใจ",
			FirstNameEN: "Anan",
			LastNameEN:  "Sukjai",
			DateOfBirth: "1992-03-15",
			PatientHN:   "HN-A-1001",
			NationalID:  "1100100000044",
			PassportID:  "AA1234567",
			PhoneNumber: "0810000004",
			Email:       "anan@example.com",
			Gender:      "M",
		},
	}
}
