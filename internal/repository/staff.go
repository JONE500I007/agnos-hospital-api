package repository

import (
	"context"
	"errors"

	"github.com/JONE500I007/agnos-hospital-api/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StaffRepository struct {
	db *pgxpool.Pool
}

func NewStaffRepository(db *pgxpool.Pool) *StaffRepository {
	return &StaffRepository{db: db}
}

func (r *StaffRepository) Create(ctx context.Context, username, passwordHash, hospital string) error {
	tag, err := r.db.Exec(ctx,
		`INSERT INTO staff (username, password_hash, hospital_id)
		 SELECT $1, $2, id FROM hospitals WHERE name = $3`,
		username, passwordHash, hospital)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.ErrUsernameTaken
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrHospitalNotFound
	}
	return nil
}

func (r *StaffRepository) GetByUsername(ctx context.Context, username string) (*model.Staff, error) {
	var s model.Staff
	err := r.db.QueryRow(ctx,
		`SELECT s.id, s.username, s.password_hash, s.hospital_id, h.name
		 FROM staff s JOIN hospitals h ON h.id = s.hospital_id
		 WHERE s.username = $1`,
		username).Scan(&s.ID, &s.Username, &s.PasswordHash, &s.HospitalID, &s.HospitalName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrStaffNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}
