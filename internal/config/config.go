package config

import "os"

type Config struct {
	DatabaseURL     string
	JWTSecret       string
	Port            string
	HISHospitalAURL string // empty: use the built-in mock HIS
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return Config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		Port:            port,
		HISHospitalAURL: os.Getenv("HIS_HOSPITAL_A_URL"),
	}
}
