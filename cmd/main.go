package main

import (
	"fmt"
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

	"net/http"
)

// @title Kir-tube API
// @version 2.0

// @host localhost:8000
// @BasePath /api

func App() (http.Handler, string) {
	conf := configs.LoadConfig()
	db := db.NewDb(conf)

	router := http.NewServeMux()

	videoRepository := video.NewVideoRepository(db)
	userProvider := channel.NewContextUserRepository(db)

	//  -- Modules -- //
	userModule := user.NewUserModule(router, user.UserModuleDeps{
		Config:          conf,
		Db:              db,
		VideoRepository: videoRepository,
		UserProvider:    userProvider,
	})

	auth.NewAuthModule(router, auth.AuthModuleDeps{
		UserService: userModule.UserService,
		Config:      conf,
		Db:          db,
	})

	video.NewVideoModule(router, video.VideoModuleDeps{
		Config:          conf,
		VideoRepository: videoRepository,
		UserProvider:    userProvider,
	})

	channel.NewChannelModule(router, channel.ChannelModuleDeps{
		VideoRepository: videoRepository,
		Config:          conf,
		Db:              db,
		UserProvider:    userProvider,
	})

	playlist.NewPlaylistModule(router, playlist.PlaylistModuleDeps{
		VideoRepository: videoRepository,
		Config:          conf,
		Db:              db,
		UserProvider:    userProvider,
	})
	studio.NewStudioModule(router, studio.StudioModuleDeps{
		VideoRepository: videoRepository,
		Config:          conf,
		Db:              db,
		UserProvider:    userProvider,
	})
	comment.NewCommentModule(router, comment.CommentModuleDeps{
		VideoRepository: videoRepository,
		Config:          conf,
		Db:              db,
		UserProvider:    userProvider,
	})

	media.NewMediaModule(router, media.MediaModuleDeps{
		Config:       conf,
		UserProvider: userProvider,
	})

	// Mount every module under a global /api prefix. StripPrefix removes it
	// before the request reaches the module routes, so the handlers keep
	// registering paths like "/comments/..." unchanged.
	apiRouter := http.NewServeMux()
	apiRouter.Handle("/api/", http.StripPrefix("/api", router))

	stack := middleware.Chain(middleware.CORS(conf.Network.ClientUrl), middleware.Logging)

	return stack(apiRouter), conf.Network.Port
}

func main() {
	app, port := App()

	server := http.Server{
		Addr:    ":" + port,
		Handler: app,
	}

	fmt.Println("📢 Server is listening on port " + port)
	server.ListenAndServe()
}
