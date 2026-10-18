package main

import (
	"context"
	"log"
	"net/http"

	"github.com/JONE500I007/agnos-hospital-api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"


	"github.com/JONE500I007/agnos-hospital-api/internal/handler"
	"github.com/JONE500I007/agnos-hospital-api/internal/repository"
	"github.com/JONE500I007/agnos-hospital-api/internal/service"

	"github.com/JONE500I007/agnos-hospital-api/internal/middleware"
)

func main() {
	cfg := config.Load()

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("cannot create db pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("cannot connect to db: %v", err)
	}

	r := gin.Default()

	staffRepo := repository.NewStaffRepository(pool)
	staffSvc := service.NewStaffService(staffRepo, cfg.JWTSecret)
	staffHandler := handler.NewStaffHandler(staffSvc)


	r.POST("/staff/login", staffHandler.Login)
	r.POST("/staff/create", staffHandler.Create)

	patientRepo := repository.NewPatientRepository(pool)
	patientSvc := service.NewPatientService(patientRepo)
	patientHandler := handler.NewPatientHandler(patientSvc)

	patients := r.Group("/patient", middleware.Auth(cfg.JWTSecret))
	patients.POST("/search", patientHandler.Search)

	patients.GET("/search/:id", patientHandler.GetByID)

	r.GET("/health", func(c *gin.Context) {
		if err := pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"message": "db down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	r.Run(":" + cfg.Port)
}