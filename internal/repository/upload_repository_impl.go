package repository

func NewUploadRepository() UploadRepository {
	return &UploadRepositoryImpl{}
}

type UploadRepositoryImpl struct{}
