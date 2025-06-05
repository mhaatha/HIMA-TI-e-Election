package web

type PresignedURLResponse struct {
	URL      string `json:"url"`
	FileName string `json:"file_name"`
}
