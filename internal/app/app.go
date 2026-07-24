// Package app wires every module into a single HTTP handler. It is kept separate
// from cmd/main.go so that tests can assemble the whole application with injected
// dependencies (a test database and an in-memory object store) instead of the
// production ones read from the environment.
package app

import (
	"net/http"

	"go/kir-tube/configs"
	"go/kir-tube/internal/auth"
	"go/kir-tube/internal/channel"
	"go/kir-tube/internal/comment"
	"go/kir-tube/internal/media"
	"go/kir-tube/internal/playlist"
	"go/kir-tube/internal/studio"
	"go/kir-tube/internal/user"
	"go/kir-tube/internal/video"
	"go/kir-tube/pkg/db"
	"go/kir-tube/pkg/middleware"
)

// Assemble builds the full application handler and returns it together with the
// port to listen on. storage may be nil, in which case the media module builds a
// MinIO-backed store from conf (the production path); tests pass an in-memory
// fake instead.
func Assemble(conf *configs.Config, database *db.Db, storage media.ObjectStorage) (http.Handler, string) {
	router := http.NewServeMux()
	// Root mux: API modules are mounted under /api (below), while public media
	// is served here at /uploads (outside the /api prefix).
	apiRouter := http.NewServeMux()

	videoRepository := video.NewVideoRepository(database)
	userProvider := channel.NewContextUserRepository(database)

	//  -- Modules -- //
	userModule := user.NewUserModule(router, user.UserModuleDeps{
		Config:          conf,
		Db:              database,
		VideoRepository: videoRepository,
		UserProvider:    userProvider,
	})

	auth.NewAuthModule(router, auth.AuthModuleDeps{
		UserService: userModule.UserService,
		Config:      conf,
		Db:          database,
	})

	video.NewVideoModule(router, video.VideoModuleDeps{
		Config:          conf,
		VideoRepository: videoRepository,
		UserProvider:    userProvider,
	})

	channel.NewChannelModule(router, channel.ChannelModuleDeps{
		VideoRepository: videoRepository,
		Config:          conf,
		Db:              database,
		UserProvider:    userProvider,
	})

	playlist.NewPlaylistModule(router, playlist.PlaylistModuleDeps{
		VideoRepository: videoRepository,
		Config:          conf,
		Db:              database,
		UserProvider:    userProvider,
	})
	studio.NewStudioModule(router, studio.StudioModuleDeps{
		VideoRepository: videoRepository,
		Config:          conf,
		Db:              database,
		UserProvider:    userProvider,
	})
	comment.NewCommentModule(router, comment.CommentModuleDeps{
		VideoRepository: videoRepository,
		Config:          conf,
		Db:              database,
		UserProvider:    userProvider,
	})

	media.NewMediaModule(router, media.MediaModuleDeps{
		Config:       conf,
		UserProvider: userProvider,
		PublicRouter: apiRouter,
		Storage:      storage,
	})

	// Mount every module under a global /api prefix. StripPrefix removes it
	// before the request reaches the module routes, so the handlers keep
	// registering paths like "/comments/..." unchanged.
	apiRouter.Handle("/api/", http.StripPrefix("/api", router))

	stack := middleware.Chain(middleware.CORS(conf.Network.ClientUrl), middleware.Logging)

	return stack(apiRouter), conf.Network.Port
}
