package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/JONE500I007/agnos-hospital-api/internal/config"
	"github.com/JONE500I007/agnos-hospital-api/internal/handler"
	"github.com/JONE500I007/agnos-hospital-api/internal/his"
	"github.com/JONE500I007/agnos-hospital-api/internal/middleware"
	"github.com/JONE500I007/agnos-hospital-api/internal/repository"
	"github.com/JONE500I007/agnos-hospital-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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

	var hospitalAHIS service.HISClient
	if cfg.HISHospitalAURL != "" {
		hospitalAHIS = his.NewHTTPClient(cfg.HISHospitalAURL, 5*time.Second)
		log.Printf("Hospital A HIS: %s", cfg.HISHospitalAURL)
	} else {
		hospitalAHIS = his.NewMockClient(his.SampleHospitalAPatients()...)
		log.Print("HIS_HOSPITAL_A_URL not set: using the built-in mock HIS for Hospital A")
	}

	r := gin.Default()

	staffRepo := repository.NewStaffRepository(pool)
	staffSvc := service.NewStaffService(staffRepo, cfg.JWTSecret)
	staffHandler := handler.NewStaffHandler(staffSvc)

	r.POST("/staff/create", staffHandler.Create)
	r.POST("/staff/login", staffHandler.Login)

	patientRepo := repository.NewPatientRepository(pool)
	patientSvc := service.NewPatientService(patientRepo, map[string]service.HISClient{
		"Hospital A": hospitalAHIS,
	})
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
