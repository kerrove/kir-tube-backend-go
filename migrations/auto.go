package main

import (
	"go/kir-tube/internal/channel"
	"go/kir-tube/internal/history"
	"go/kir-tube/internal/playlist"
	"go/kir-tube/internal/user"
	"go/kir-tube/internal/video"

	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err.Error())
	}
	db, err := gorm.Open(postgres.Open(os.Getenv("DSN")), &gorm.Config{})

	if err != nil {
		panic(err.Error())
	}

	db.AutoMigrate(&video.Video{}, &history.WatchHistory{}, &playlist.Playlist{}, &channel.Channel{}, &user.User{}, &video.VideoComment{}, &video.VideoLike{})

}
