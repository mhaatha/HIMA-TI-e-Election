package web

type WebSuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
