package model

import "errors"

var (
	ErrUsernameTaken    = errors.New("username already exists")
	ErrHospitalNotFound = errors.New("hospital not found")
	ErrStaffNotFound    = errors.New("staff not found")

	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrPatientNotFound    = errors.New("patient not found")
	ErrHISUnavailable     = errors.New("hospital information system unavailable")
)
