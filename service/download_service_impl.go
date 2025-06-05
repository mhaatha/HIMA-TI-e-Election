package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	envConfig "github.com/mhaatha/HIMA-TI-e-Election/config"
	myErrors "github.com/mhaatha/HIMA-TI-e-Election/errors"
	"github.com/mhaatha/HIMA-TI-e-Election/model/web"
	"github.com/mhaatha/HIMA-TI-e-Election/repository"
)

func NewDownloadService(downloadRepository repository.DownloadRepository, envConfig *envConfig.Config, db *pgxpool.Pool, validate *validator.Validate) DownloadService {
	return &DownloadServiceImpl{
		DownloadRepository: downloadRepository,
		EnvConfig:          envConfig,
		DB:                 db,
		Validate:           validate,
	}
}

type DownloadServiceImpl struct {
	DownloadRepository repository.DownloadRepository
	EnvConfig          *envConfig.Config
	DB                 *pgxpool.Pool
	Validate           *validator.Validate
}

func (service *DownloadServiceImpl) CreatePresignedURL(ctx context.Context, fileName string) (web.PresignedURLResponse, error) {
	// Validate the fileName, it should can't exceed 4 digit
	if len(fileName) != 4 {
		return web.PresignedURLResponse{}, myErrors.NewAppError(
			http.StatusBadRequest,
			"Invalid file name",
			"File name should be 4 digits",
			fmt.Errorf("file name is not 4 digits: %v digits length", len(fileName)),
		)
	}

	// Validate the fileName, it should be converted to int
	_, err := strconv.Atoi(fileName)
	if err != nil {
		// Filename must be a valid number
		if errors.Is(err, strconv.ErrSyntax) {
			return web.PresignedURLResponse{}, myErrors.NewAppError(
				http.StatusBadRequest,
				"Invalid request payload",
				"Filename must be a number and can't contain characters",
				fmt.Errorf("file name is not a number: %v", err),
			)
		}

		return web.PresignedURLResponse{}, myErrors.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("file to convert filename to int: %v", err),
		)
	}

	// Load S3 default config
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(service.EnvConfig.S3AccessKeyId, service.EnvConfig.S3SecretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return web.PresignedURLResponse{}, myErrors.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("%w: %v", myErrors.ErrLoadDefaultConfig, err),
		)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(service.EnvConfig.S3URL)
	})

	// Read the log local file
	logFilePath, err := envConfig.GetLogFilePathByYear(fileName)
	if err != nil {
		return web.PresignedURLResponse{}, myErrors.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("error getting log file path: %v", err),
		)
	}

	file, err := os.Open(logFilePath)
	if err != nil {
		return web.PresignedURLResponse{}, myErrors.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to process your request due to an unexpected error. Please try again later.",
			fmt.Errorf("error opening log file: %v", err),
		)
	}

	// Upload to Object Storage
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(service.EnvConfig.S3Bucket),
		Key:    aws.String(fmt.Sprintf("logs/vote_%s.log", fileName)),
		Body:   file,
	})
	if err != nil {
		return web.PresignedURLResponse{}, myErrors.NewAppError(
			http.StatusInternalServerError,
			"Failed to upload log",
			"Server failed to upload the log file to storage",
			fmt.Errorf("failed to upload file to S3: %v", err),
		)
	}

	// Create the presigned URL to get the uploaded log file
	presignClient := s3.NewPresignClient(client)

	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(service.EnvConfig.S3Bucket),
		Key:    aws.String(fmt.Sprintf("logs/vote_%s.log", fileName)),
	}, s3.WithPresignExpires(5*time.Minute))
	if err != nil {
		return web.PresignedURLResponse{}, myErrors.NewAppError(
			http.StatusInternalServerError,
			"Failed to generate download link",
			"Server failed to generate presigned URL for log download",
			fmt.Errorf("failed to presign GET URL: %v", err),
		)
	}

	return web.PresignedURLResponse{
		URL:      presignResult.URL,
		FileName: fmt.Sprintf("vote_%s.log", fileName),
	}, nil
}
