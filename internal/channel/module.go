package channel

import (
	"go/kir-tube/configs"
	"go/kir-tube/pkg/db"
	"go/kir-tube/pkg/di"
	"net/http"
)

type ChannelModuleDeps struct {
	Config          *configs.Config
	Db              *db.Db
	VideoRepository di.IChannelVideoRepository
}
type ChannelModule struct {
	ChannelService *ChannelService
}

func NewChannelModule(router *http.ServeMux, deps ChannelModuleDeps) {
	channelRepository := NewChannelRepository(deps.Db, deps.VideoRepository)

	channelService :=
		NewChannelService(&ChannelServiceDeps{
			ChannelRepository: channelRepository,
		})

	NewChannelHandler(router, ChannelHandlerDeps{
		ChannelService: channelService,
		Config:         deps.Config,
	})

}
