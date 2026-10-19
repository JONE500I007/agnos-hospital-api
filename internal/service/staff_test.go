package service

import (
	"context"
	"errors"
	"testing"

	"github.com/JONE500I007/agnos-hospital-api/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const testSecret = "unit-test-secret"

type fakeStaffStore struct {
	createErr error
	staff     *model.Staff
	getErr    error

	gotUsername string
	gotHash     string
	gotHospital string
}

func (f *fakeStaffStore) Create(ctx context.Context, username, passwordHash, hospital string) error {
	f.gotUsername = username
	f.gotHash = passwordHash
	f.gotHospital = hospital
	return f.createErr
}

func (f *fakeStaffStore) GetByUsername(ctx context.Context, username string) (*model.Staff, error) {
	return f.staff, f.getErr
}

func TestStaffServiceCreate(t *testing.T) {
	t.Run("stores a bcrypt hash, not the plain password", func(t *testing.T) {
		store := &fakeStaffStore{}
		svc := NewStaffService(store, testSecret)

		if err := svc.Create(context.Background(), "nurse1", "pass1234", "Hospital A"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if store.gotHash == "pass1234" {
			t.Fatal("password was stored in plain text")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(store.gotHash), []byte("pass1234")); err != nil {
			t.Errorf("stored hash does not match password: %v", err)
		}
		if store.gotUsername != "nurse1" || store.gotHospital != "Hospital A" {
			t.Errorf("store got username=%q hospital=%q", store.gotUsername, store.gotHospital)
		}
	})

	t.Run("passes store errors through", func(t *testing.T) {
		svc := NewStaffService(&fakeStaffStore{createErr: model.ErrUsernameTaken}, testSecret)
		err := svc.Create(context.Background(), "nurse1", "pass1234", "Hospital A")
		if !errors.Is(err, model.ErrUsernameTaken) {
			t.Errorf("error = %v, want ErrUsernameTaken", err)
		}
	})
}

func TestStaffServiceLogin(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("pass1234"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	staff := &model.Staff{ID: 1, Username: "nurse1", PasswordHash: string(hash), HospitalID: 3}

	t.Run("success returns a token carrying the hospital", func(t *testing.T) {
		svc := NewStaffService(&fakeStaffStore{staff: staff}, testSecret)

		token, err := svc.Login(context.Background(), "nurse1", "pass1234")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		parsed, err := jwt.Parse(token,
			func(*jwt.Token) (any, error) { return []byte(testSecret), nil },
			jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !parsed.Valid {
			t.Fatalf("token is not valid: %v", err)
		}
		claims := parsed.Claims.(jwt.MapClaims)
		if claims["hospital_id"] != float64(3) {
			t.Errorf("hospital_id = %v, want 3", claims["hospital_id"])
		}
		if claims["staff_id"] != float64(1) {
			t.Errorf("staff_id = %v, want 1", claims["staff_id"])
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		svc := NewStaffService(&fakeStaffStore{staff: staff}, testSecret)
		token, err := svc.Login(context.Background(), "nurse1", "wrongpass")
		if !errors.Is(err, model.ErrInvalidCredentials) {
			t.Errorf("error = %v, want ErrInvalidCredentials", err)
		}
		if token != "" {
			t.Error("token must be empty on failure")
		}
	})

	t.Run("unknown user gives the same error as wrong password", func(t *testing.T) {
		svc := NewStaffService(&fakeStaffStore{getErr: model.ErrStaffNotFound}, testSecret)
		_, err := svc.Login(context.Background(), "ghost", "pass1234")
		if !errors.Is(err, model.ErrInvalidCredentials) {
			t.Errorf("error = %v, want ErrInvalidCredentials", err)
		}
	})

	t.Run("store failure is not reported as bad credentials", func(t *testing.T) {
		dbErr := errors.New("db down")
		svc := NewStaffService(&fakeStaffStore{getErr: dbErr}, testSecret)
		_, err := svc.Login(context.Background(), "nurse1", "pass1234")
		if !errors.Is(err, dbErr) {
			t.Errorf("error = %v, want %v", err, dbErr)
		}
	})
}