package service

import (
	"context"
	"errors"
	"time"

	"github.com/JONE500I007/agnos-hospital-api/internal/model"
	"github.com/golang-jwt/jwt/v5"

	"golang.org/x/crypto/bcrypt"
)

type StaffStore interface {
	Create(ctx context.Context, username, passwordHash, hospital string) error
	GetByUsername(ctx context.Context, username string) (*model.Staff, error)
}

type StaffService struct {
	store     StaffStore
	jwtSecret string
}

func NewStaffService(store StaffStore, jwtSecret string) *StaffService {
	return &StaffService{store: store, jwtSecret: jwtSecret}
}

func (s *StaffService) Create(ctx context.Context, username, password, hospital string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.store.Create(ctx, username, string(hash), hospital)
}

func (s *StaffService) Login(ctx context.Context, username, password, hospital string) (string, error) {
	staff, err := s.store.GetByUsername(ctx, username)
	if errors.Is(err, model.ErrStaffNotFound) {
		return "", model.ErrInvalidCredentials
	}
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(password)); err != nil {
		return "", model.ErrInvalidCredentials
	}

	if staff.HospitalName != hospital {
		return "", model.ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"staff_id":    staff.ID,
		"hospital_id": staff.HospitalID,
		"exp":         time.Now().Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtSecret))
}
