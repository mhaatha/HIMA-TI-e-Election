package web

type LoginRequest struct {
	NIM      string `json:"nim" validate:"omitempty,min=4,max=14"`
	Password string `json:"password" validate:"required,min=6,max=100"`
}
