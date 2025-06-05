package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	S3AccessKeyId     string
	S3SecretAccessKey string
	S3Bucket          string
	S3URL             string

	FonnteAPIKey string
}

func LoadConfig() (*Config, error) {
	envFilePath, err := GetEnvFilePath()
	if err != nil {
		return nil, fmt.Errorf("error getting current directory: %v", err)
	}

	err = godotenv.Load(envFilePath)
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	return &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),

		S3AccessKeyId:     os.Getenv("S3_ACCESS_KEY_ID"),
		S3SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
		S3Bucket:          os.Getenv("S3_BUCKET"),
		S3URL:             os.Getenv("S3_URL"),

		FonnteAPIKey: os.Getenv("FONNTE_API_KEY"),
	}, nil
}

func GetEnvFilePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	exeDir := filepath.Dir(exePath)
	envPath := filepath.Join(exeDir, "..", "..", ".env")
	return envPath, nil
}
