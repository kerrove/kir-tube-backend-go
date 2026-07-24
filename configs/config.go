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
	Storage StorageConfig
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

// StorageConfig holds the MinIO/S3 connection settings for the media module.
type StorageConfig struct {
	Endpoint  string // host:port without scheme, e.g. localhost:9000
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
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
		Storage: StorageConfig{
			Endpoint:  os.Getenv("MINIO_ENDPOINT"),
			AccessKey: os.Getenv("MINIO_ROOT_USER"),
			SecretKey: os.Getenv("MINIO_ROOT_PASSWORD"),
			Bucket:    getEnvDefault("MINIO_BUCKET", "kir-tube"),
			UseSSL:    parseBool(os.Getenv("MINIO_USE_SSL")),
		},
	}
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseBool(v string) bool {
	b, err := strconv.ParseBool(v)
	return err == nil && b
}
