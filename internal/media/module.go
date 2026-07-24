package media

import (
	"log"
	"net/http"

	"go/kir-tube/configs"
	"go/kir-tube/pkg/di"
)

const defaultPublicBaseURL = "/uploads"

type MediaModuleDeps struct {
	Config       *configs.Config
	UserProvider di.IUserProvider
	// PublicRouter is the root mux on which static media is served at /uploads.
	// It is kept separate from the API router so uploads live outside /api.
	PublicRouter *http.ServeMux
	// Storage lets callers inject an object store (e.g. an in-memory fake in
	// tests). When nil the module builds a MinIO-backed store from Config, which
	// is the production path.
	Storage ObjectStorage
}
type MediaModule struct {
	MediaService *MediaService
}

func NewMediaModule(router *http.ServeMux, deps MediaModuleDeps) *MediaModule {
	storage := deps.Storage
	if storage == nil {
		minioStorage, err := NewMinioStorage(deps.Config.Storage)
		if err != nil {
			log.Fatalf("media: failed to init object storage: %v", err)
		}
		storage = minioStorage
	}

	mediaService := NewMediaService(&MediaServiceDeps{
		Storage:       storage,
		PublicBaseURL: defaultPublicBaseURL,
	})

	NewMediaHandler(router, deps.PublicRouter, MediaHandlerDeps{
		MediaService: mediaService,
		Config:       deps.Config,
		UserProvider: deps.UserProvider,
		Storage:      storage,
	})

	return &MediaModule{MediaService: mediaService}
}
