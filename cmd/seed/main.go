package main

import (
	"flag"
	"fmt"
	"os"

	"go/kir-tube/configs"
	"go/kir-tube/pkg/db"
	"go/kir-tube/seeder"
)

func main() {
	clean := flag.Bool("clean", true, "remove previously seeded data before seeding")
	flag.Parse()

	conf := configs.LoadConfig()
	database := db.NewDb(conf)

	report, err := seeder.NewSeeder(database).Run(*clean)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка при заполнении базы данных:", err)
		os.Exit(1)
	}

	fmt.Printf(
		"Заполнение базы данных завершено успешно: каналов — %d, видео — %d, тегов — %d.\n",
		report.Channels, report.Videos, report.Tags,
	)
}
