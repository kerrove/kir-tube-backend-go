package main

import (
	"fmt"
	"go/kir-tube/configs"
	"go/kir-tube/internal/app"
	"go/kir-tube/pkg/db"
	"log"
	"time"

	"net/http"
)

// @title Kir-tube API
// @version 2.0

// @host localhost:8000
// @BasePath /api

func App() (http.Handler, string) {
	conf := configs.LoadConfig()
	database := db.NewDb(conf)

	// nil storage → the media module builds the production MinIO store from conf.
	return app.Assemble(conf, database, nil)
}

func main() {
	app, port := App()

	server := http.Server{
		Addr:              ":" + port,
		Handler:           app,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	fmt.Println("📢 Server is listening on port " + port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("🛑 Server failed: %v", err)
	}
}
