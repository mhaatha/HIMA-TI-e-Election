package web

type WebFailedResponse struct {
	Message string      `json:"message"`
	Details interface{} `json:"details"`
}
