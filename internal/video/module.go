package video

import (
	"go/kir-tube/configs"
	"net/http"
)

type VideoModuleDeps struct {
	Config          *configs.Config
	VideoRepository IVideoRepository
}

func NewVideoModule(router *http.ServeMux, deps VideoModuleDeps) {
	videoService :=
		NewVideoService(&VideoServiceDeps{
			VideoRepository: deps.VideoRepository,
		})
	NewVideoHandler(router, VideoHandlerDeps{
		VideoService: videoService,
		Config:       deps.Config,
	})
}
