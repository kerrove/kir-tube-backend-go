package main

import (
	"fmt"
	"go/kir-tube/configs"
	"go/kir-tube/internal/auth"
	"go/kir-tube/internal/channel"
	"go/kir-tube/internal/playlist"
	"go/kir-tube/internal/studio"
	"go/kir-tube/internal/user"
	"go/kir-tube/internal/video"
	"go/kir-tube/pkg/db"
	"go/kir-tube/pkg/middleware"

	"net/http"
)

func App() (http.Handler, string) {
	conf := configs.LoadConfig()
	db := db.NewDb(conf)

	router := http.NewServeMux()

	videoRepository := video.NewVideoRepository(db)
	// userProvider loads the full authenticated user (with their channel) for the
	// auth middleware (di.IUserProvider). It lives in the channel package because
	// it reads both the user and channel tables. Shared by every module whose
	// routes are behind IsAuthed / MaybeAuthed.
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

	stack := middleware.Chain(middleware.CORS, middleware.Logging)

	return stack(router), conf.Network.Port
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
