package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	DBURL string

	S3AccessKeyId     string
	S3SecretAccessKey string
	S3Bucket          string
	S3URL             string

	FonnteAPIKey  string
	FonnteSendURL string

	FrontendURL string
}

func LoadConfig() (*Config, error) {
	envFilePath, err := GetEnvFilePath()
	if err != nil {
		return nil, fmt.Errorf("getting current directory: %w", err)
	}

	err = godotenv.Load(envFilePath)
	if err != nil {
		return nil, fmt.Errorf("loading .env file: %w", err)
	}

	return &Config{
		DBURL: os.Getenv("DB_URL"),

		S3AccessKeyId:     os.Getenv("S3_ACCESS_KEY_ID"),
		S3SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
		S3Bucket:          os.Getenv("S3_BUCKET"),
		S3URL:             os.Getenv("S3_URL"),

		FonnteAPIKey:  os.Getenv("FONNTE_API_KEY"),
		FonnteSendURL: os.Getenv("FONNTE_SEND_URL"),

		FrontendURL: os.Getenv("FRONTEND_URL"),
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
