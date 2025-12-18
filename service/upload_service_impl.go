package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	envConfig "github.com/mhaatha/HIMA-TI-e-Election/config"
	appError "github.com/mhaatha/HIMA-TI-e-Election/errors"
	"github.com/mhaatha/HIMA-TI-e-Election/model/web"
	"github.com/mhaatha/HIMA-TI-e-Election/repository"
)

func NewUploadService(uploadRepository repository.UploadRepository, envConfig *envConfig.Config, db *pgxpool.Pool, validate *validator.Validate) UploadService {
	return &UploadServiceImpl{
		UploadRepository: uploadRepository,
		EnvConfig:        envConfig,
		DB:               db,
		Validate:         validate,
	}
}

type UploadServiceImpl struct {
	UploadRepository repository.UploadRepository
	EnvConfig        *envConfig.Config
	DB               *pgxpool.Pool
	Validate         *validator.Validate
}

func (service *UploadServiceImpl) CreatePresignedURL(ctx context.Context, fileName string) (web.PresignedURLResponse, error) {
	// Load S3 default config
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(service.EnvConfig.S3AccessKeyId, service.EnvConfig.S3SecretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return web.PresignedURLResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrLoadDefaultConfig, err),
		)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(service.EnvConfig.S3URL)
	})

	// Presigned URL client
	presignClient := s3.NewPresignClient(client)

	// Create presigned URL for PutObject in 5 minutes
	presignResult, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(service.EnvConfig.S3Bucket),
		Key:    aws.String(fileName),
	}, s3.WithPresignExpires(5*time.Minute))
	if err != nil {
		return web.PresignedURLResponse{}, appError.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", appError.ErrCreatePresignedPut, err),
		)
	}

	return web.PresignedURLResponse{
		URL:      presignResult.URL,
		FileName: fileName,
	}, nil
}
