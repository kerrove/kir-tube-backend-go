# Конфигурация

Конфигурация читается из переменных окружения при старте. Файл `.env`
загружается через `godotenv` (`configs/config.go`, `LoadConfig`). Если `.env` не
найден — приложение продолжает работать на значениях из окружения (или пустых).

> **Секреты не коммитить.** `.env` должен быть в `.gitignore`. `JWT_SECRET`,
> пароли БД и MinIO — это секреты.

## Переменные окружения

### Сеть и приложение

| Переменная | Пример | Куда идёт | Описание |
|------------|--------|-----------|----------|
| `PORT` | `8000` | `Network.Port` | Порт HTTP-сервера |
| `DOMAIN` | `localhost` | `Network.Domain` | Домен приложения |
| `CLIENT_URL` | `http://localhost:3000` | `Network.ClientUrl` | Origin фронтенда для CORS (разрешается только он) |
| `MODE` | `development` | — | Режим окружения (информационно) |

### Аутентификация

| Переменная | Пример | Куда идёт | Описание |
|------------|--------|-----------|----------|
| `JWT_SECRET` | `s+9ntw...=` | `Auth.Secret` | Секрет для подписи JWT (HS256). **Секрет.** |
| `SECURE_COOKIES` | `false` | `Auth.SecureCookies` | Флаг `Secure` у refresh-cookie. В проде за HTTPS — `true` |

### База данных

| Переменная | Пример | Куда идёт | Описание |
|------------|--------|-----------|----------|
| `DSN` | `host=db user=postgres password=... dbname=kir-tube port=5432 sslmode=disable` | `Db.Dsn` | Строка подключения к PostgreSQL (GORM/pgx) |
| `TEST_DSN` | `host=db ... dbname=kir-tube_test ...` | — | Строка подключения тестовой БД (используется интеграционными тестами) |

Дополнительные переменные PostgreSQL используются **контейнером** `postgres` в
`docker-compose.yaml` (через `env_file`), но приложением напрямую **не читаются**
(оно берёт всё из `DSN`):

| Переменная | Описание |
|------------|----------|
| `POSTGRES_USER` | Пользователь БД, создаётся контейнером |
| `POSTGRES_PASSWORD` | Пароль БД |
| `POSTGRES_DB` | Имя создаваемой БД |
| `POSTGRES_HOST` | Хост (для контейнера/справочно; в приложении хост берётся из `DSN`) |
| `POSTGRES_PORT` | Порт |
| `POSTGRES_DATA` | Путь к данным внутри контейнера |

> **Согласованность `DSN` ↔ Docker.** `host=` в `DSN` — это имя, по которому
> приложение находит БД. Внутри Compose это имя сервиса БД (`db`), при локальном
> запуске — `localhost`. Значения `POSTGRES_*` (например `POSTGRES_HOST=localhost`)
> в подключении приложения не участвуют и могут вводить в заблуждение — источник
> истины один: `DSN`.

### Объектное хранилище (MinIO / S3)

| Переменная | Пример | Куда идёт | Описание |
|------------|--------|-----------|----------|
| `MINIO_ENDPOINT` | `minio:9000` | `Storage.Endpoint` | Хост:порт без схемы |
| `MINIO_ROOT_USER` | `minio` | `Storage.AccessKey` | Access key |
| `MINIO_ROOT_PASSWORD` | `55Kirill55` | `Storage.SecretKey` | Secret key. **Секрет.** |
| `MINIO_BUCKET` | `kir-tube` | `Storage.Bucket` | Бакет для медиа (по умолчанию `kir-tube`) |
| `MINIO_USE_SSL` | `false` | `Storage.UseSSL` | Использовать HTTPS до MinIO |

При старте медиа-модуль подключается к MinIO и **создаёт бакет**, если его нет.

## Как это ложится в `Config`

```go
type Config struct {
    Db      DbConfig      // Dsn
    Auth    AuthConfig    // Secret, SecureCookies
    Network NetworkConfig // Port, Domain, ClientUrl
    Storage StorageConfig // Endpoint, AccessKey, SecretKey, Bucket, UseSSL
}
```

- `parseBool` трактует значение как `true` только при корректном булевом
  представлении (`true`/`1`/…); иначе `false`.
- `MINIO_BUCKET` имеет дефолт `kir-tube` (`getEnvDefault`).
