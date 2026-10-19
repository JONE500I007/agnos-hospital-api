package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/JONE500I007/agnos-hospital-api/internal/middleware"
	"github.com/JONE500I007/agnos-hospital-api/internal/model"
	"github.com/gin-gonic/gin"
)

type PatientService interface {
	Search(ctx context.Context, hospitalID int, f model.PatientSearch) ([]model.Patient, error)
	FindByID(ctx context.Context, hospitalID int, id string) (*model.Patient, error)
}

type PatientHandler struct {
	svc PatientService
}

func NewPatientHandler(svc PatientService) *PatientHandler {
	return &PatientHandler{svc: svc}
}

func (h *PatientHandler) Search(c *gin.Context) {
	var f model.PatientSearch
	if err := c.ShouldBindJSON(&f); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if f.DateOfBirth != "" {
		if _, err := time.Parse("2006-01-02", f.DateOfBirth); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "date_of_birth must be YYYY-MM-DD"})
			return
		}
	}

	patients, err := h.svc.Search(c.Request.Context(), c.GetInt(middleware.HospitalIDKey), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, patients)
}

func (h *PatientHandler) GetByID(c *gin.Context) {
	patient, err := h.svc.FindByID(c.Request.Context(), c.GetInt(middleware.HospitalIDKey), c.Param("id"))
	switch {
	case errors.Is(err, model.ErrPatientNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, model.ErrHISUnavailable):
		c.JSON(http.StatusBadGateway, gin.H{"error": model.ErrHISUnavailable.Error()})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	default:
		c.JSON(http.StatusOK, patient)
	}
}
