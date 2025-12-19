package service

import (
	"context"

	"github.com/mhaatha/HIMA-TI-e-Election/internal/model/web"
)

type DownloadService interface {
	CreatePresignedURL(ctx context.Context, fileName string) (web.PresignedURLResponse, error)
}
