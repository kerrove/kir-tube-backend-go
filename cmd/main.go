package main

import (
	"fmt"
	"go/kir-tube/configs"
	"go/kir-tube/internal/auth"
	"go/kir-tube/internal/channel"
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

	//  -- Modules -- //
	userModule := user.NewUserModule(router, user.UserModuleDeps{
		Config:          conf,
		Db:              db,
		VideoRepository: videoRepository,
	})

	auth.NewAuthModule(router, auth.AuthModuleDeps{
		UserService: userModule.UserService,
		Config:      conf,
		Db:          db,
	})

	video.NewVideoModule(router, video.VideoModuleDeps{
		Config:          conf,
		VideoRepository: videoRepository,
	})

	channel.NewChannelModule(router, channel.ChannelModuleDeps{
		VideoRepository: videoRepository,
		Config:          conf,
		Db:              db,
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
