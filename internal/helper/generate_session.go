package helper

import (
	"encoding/base64"

	"github.com/google/uuid"
)

// Base64SessionId returns a base64 encoded uuid
func Base64SessionId() string {
	uuid := uuid.New()
	return base64.StdEncoding.EncodeToString(uuid[:])
}
