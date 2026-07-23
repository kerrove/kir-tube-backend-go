package media

type MediaServiceDeps struct{}

type MediaService struct{}

func NewMediaService(deps *MediaServiceDeps) *MediaService {
	return &MediaService{}
}
