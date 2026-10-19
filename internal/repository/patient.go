package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/JONE500I007/agnos-hospital-api/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const patientColumns = `
	COALESCE(first_name_th, ''), COALESCE(middle_name_th, ''), COALESCE(last_name_th, ''),
	COALESCE(first_name_en, ''), COALESCE(middle_name_en, ''), COALESCE(last_name_en, ''),
	COALESCE(to_char(date_of_birth, 'YYYY-MM-DD'), ''), COALESCE(patient_hn, ''),
	COALESCE(national_id, ''), COALESCE(passport_id, ''), COALESCE(phone_number, ''),
	COALESCE(email, ''), COALESCE(gender::text, '')`

type PatientRepository struct {
	db *pgxpool.Pool
}

func NewPatientRepository(db *pgxpool.Pool) *PatientRepository {
	return &PatientRepository{db: db}
}

type scanner interface {
	Scan(dest ...any) error
}

func scanPatient(s scanner) (model.Patient, error) {
	var p model.Patient
	err := s.Scan(&p.FirstNameTH, &p.MiddleNameTH, &p.LastNameTH,
		&p.FirstNameEN, &p.MiddleNameEN, &p.LastNameEN,
		&p.DateOfBirth, &p.PatientHN, &p.NationalID, &p.PassportID,
		&p.PhoneNumber, &p.Email, &p.Gender)
	return p, err
}

func (r *PatientRepository) Search(ctx context.Context, hospitalID int, f model.PatientSearch) ([]model.Patient, error) {
	conds := []string{"hospital_id = $1"}
	args := []any{hospitalID}

	add := func(cond, val string) {
		args = append(args, val)
		conds = append(conds, fmt.Sprintf(cond, len(args)))
	}

	if f.NationalID != "" {
		add("national_id = $%[1]d", f.NationalID)
	}
	if f.PassportID != "" {
		add("passport_id = $%[1]d", f.PassportID)
	}
	if f.FirstName != "" {
		add("(first_name_th ILIKE $%[1]d OR first_name_en ILIKE $%[1]d)", "%"+f.FirstName+"%")
	}
	if f.MiddleName != "" {
		add("(middle_name_th ILIKE $%[1]d OR middle_name_en ILIKE $%[1]d)", "%"+f.MiddleName+"%")
	}
	if f.LastName != "" {
		add("(last_name_th ILIKE $%[1]d OR last_name_en ILIKE $%[1]d)", "%"+f.LastName+"%")
	}
	if f.DateOfBirth != "" {
		add("date_of_birth = $%[1]d::date", f.DateOfBirth)
	}
	if f.PhoneNumber != "" {
		add("phone_number = $%[1]d", f.PhoneNumber)
	}
	if f.Email != "" {
		add("LOWER(email) = LOWER($%[1]d)", f.Email)
	}

	query := "SELECT " + patientColumns +
		" FROM patients WHERE " + strings.Join(conds, " AND ") + " ORDER BY id"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	patients := []model.Patient{}
	for rows.Next() {
		p, err := scanPatient(rows)
		if err != nil {
			return nil, err
		}
		patients = append(patients, p)
	}
	return patients, rows.Err()
}

func (r *PatientRepository) FindByID(ctx context.Context, hospitalID int, id string) (*model.Patient, error) {
	query := "SELECT " + patientColumns +
		" FROM patients WHERE hospital_id = $1 AND (national_id = $2 OR passport_id = $2) LIMIT 1"

	p, err := scanPatient(r.db.QueryRow(ctx, query, hospitalID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrPatientNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PatientRepository) HospitalName(ctx context.Context, hospitalID int) (string, error) {
	var name string
	err := r.db.QueryRow(ctx, "SELECT name FROM hospitals WHERE id = $1", hospitalID).Scan(&name)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", model.ErrHospitalNotFound
	}
	return name, err
}

func (r *PatientRepository) Save(ctx context.Context, hospitalID int, p model.Patient) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO patients (hospital_id, first_name_th, middle_name_th, last_name_th,
			first_name_en, middle_name_en, last_name_en, date_of_birth, patient_hn,
			national_id, passport_id, phone_number, email, gender)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, '')::date, $9,
			NULLIF($10, ''), NULLIF($11, ''), $12, $13, NULLIF($14, ''))`,
		hospitalID, p.FirstNameTH, p.MiddleNameTH, p.LastNameTH,
		p.FirstNameEN, p.MiddleNameEN, p.LastNameEN, p.DateOfBirth, p.PatientHN,
		p.NationalID, p.PassportID, p.PhoneNumber, p.Email, p.Gender)
	return err
}
