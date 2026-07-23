package video

import (
	"go/kir-tube/configs"
	"go/kir-tube/pkg/di"
	"net/http"
)

type VideoModuleDeps struct {
	Config          *configs.Config
	VideoRepository IVideoRepository
	UserProvider    di.IUserProvider
}

func NewVideoModule(router *http.ServeMux, deps VideoModuleDeps) {
	videoService :=
		NewVideoService(&VideoServiceDeps{
			VideoRepository: deps.VideoRepository,
		})
	NewVideoHandler(router, VideoHandlerDeps{
		VideoService: videoService,
		Config:       deps.Config,
		UserProvider: deps.UserProvider,
	})
}
