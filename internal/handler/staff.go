package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/JONE500I007/agnos-hospital-api/internal/model"
	"github.com/gin-gonic/gin"
)

type StaffService interface {
	Create(ctx context.Context, username, password, hospital string) error
	Login(ctx context.Context, username, password, hospital string) (string, error)
}

type StaffHandler struct {
	svc StaffService
}

func NewStaffHandler(svc StaffService) *StaffHandler {
	return &StaffHandler{svc: svc}
}

type createStaffRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Hospital string `json:"hospital" binding:"required"`
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Hospital string `json:"hospital" binding:"required"`
}

func (h *StaffHandler) Create(c *gin.Context) {
	var req createStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := h.svc.Create(c.Request.Context(), req.Username, req.Password, req.Hospital)
	switch {
	case errors.Is(err, model.ErrUsernameTaken):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, model.ErrHospitalNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	default:
		c.JSON(http.StatusCreated, gin.H{"message": "staff created"})
	}
}

func (h *StaffHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	token, err := h.svc.Login(c.Request.Context(), req.Username, req.Password, req.Hospital)
	switch {
	case errors.Is(err, model.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	default:
		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}
