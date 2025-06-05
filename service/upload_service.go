package service

import (
	"context"

	"github.com/mhaatha/HIMA-TI-e-Election/model/web"
)

type UploadService interface {
	CreatePresignedURL(ctx context.Context, fileName string) (web.PresignedURLResponse, error)
}
