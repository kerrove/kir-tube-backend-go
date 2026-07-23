package configs

import (
	"log"
	"os"
	"strconv"

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
	Secret        string
	SecureCookies bool
}

type NetworkConfig struct {
	Port      string
	Domain    string
	ClientUrl string
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
			Secret:        os.Getenv("JWT_SECRET"),
			SecureCookies: parseBool(os.Getenv("SECURE_COOKIES")),
		},
		Network: NetworkConfig{
			Port:      os.Getenv("PORT"),
			Domain:    os.Getenv("DOMAIN"),
			ClientUrl: os.Getenv("CLIENT_URL"),
		},
	}
}

func parseBool(v string) bool {
	b, err := strconv.ParseBool(v)
	return err == nil && b
}
