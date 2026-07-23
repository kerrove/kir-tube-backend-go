package playlist

import (
	"go/kir-tube/configs"
	"go/kir-tube/pkg/db"
	"go/kir-tube/pkg/di"
	"net/http"
)

type PlaylistModuleDeps struct {
	Config          *configs.Config
	Db              *db.Db
	VideoRepository di.IPlaylistVideoRepository
	UserProvider    di.IUserProvider
}
type PlaylistModule struct {
	PlaylistService *PlaylistService
}

func NewPlaylistModule(router *http.ServeMux, deps PlaylistModuleDeps) {
	playlistRepository := NewPlaylistRepository(deps.Db, deps.VideoRepository)

	playlistService :=
		NewPlaylistService(&PlaylistServiceDeps{
			PlaylistRepository: playlistRepository,
		})

	NewPlaylistHandler(router, PlaylistHandlerDeps{
		PlaylistService: playlistService,
		Config:          deps.Config,
		UserProvider:    deps.UserProvider,
	})

}
