package media

import (
	"net/http"

	"go/kir-tube/configs"
	"go/kir-tube/pkg/di"
)

const defaultOutputDir = "uploads"

type MediaModuleDeps struct {
	Config       *configs.Config
	UserProvider di.IUserProvider
	OutputDir    string
}
type MediaModule struct {
	MediaService *MediaService
}

func NewMediaModule(router *http.ServeMux, deps MediaModuleDeps) *MediaModule {
	outputDir := deps.OutputDir
	if outputDir == "" {
		outputDir = defaultOutputDir
	}

	mediaService := NewMediaService(&MediaServiceDeps{OutputDir: outputDir})

	NewMediaHandler(router, MediaHandlerDeps{
		MediaService: mediaService,
		Config:       deps.Config,
		UserProvider: deps.UserProvider,
		OutputDir:    outputDir,
	})

	return &MediaModule{MediaService: mediaService}
}
