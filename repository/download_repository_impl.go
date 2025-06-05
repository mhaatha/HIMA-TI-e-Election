package repository

func NewDownloadRepository() DownloadRepository {
	return &DownloadRepositoryImpl{}
}

type DownloadRepositoryImpl struct{}
