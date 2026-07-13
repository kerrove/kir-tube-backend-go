package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Db      DbConfig
	Auth    AuthConfig
	Network NetworkConfig
}

type DbConfig struct {
	Dsn string
}
type AuthConfig struct {
	Secret string
}
type NetworkConfig struct {
	Port   string
	Domain string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using default config")
	}
	return &Config{
		Db: DbConfig{
			Dsn: os.Getenv("DSN"),
		},
		Auth: AuthConfig{
			Secret: os.Getenv("JWT_SECRET"),
		},
		Network: NetworkConfig{
			Port:   os.Getenv("PORT"),
			Domain: os.Getenv("DOMAIN"),
		},
	}
}
