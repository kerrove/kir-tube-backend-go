# Начало работы

## Требования

- **Go** 1.26.3+
- **PostgreSQL** (или Docker)
- **MinIO** (или Docker) — объектное хранилище для медиа
- **ffmpeg** и **ffprobe** в `PATH` — нужны для транскодирования видео при загрузке
- **Docker** + **Docker Compose** (для контейнерного запуска)

## 1. Конфигурация

Создайте `.env` в корне проекта. Полный список переменных — в
[configuration.md](./configuration.md). Минимальный набор для локального запуска:

```dotenv
JWT_SECRET="<случайная-строка-base64>"
PORT=8000
DOMAIN="localhost"
CLIENT_URL="http://localhost:3000"

POSTGRES_USER="postgres"
POSTGRES_PASSWORD="postgres"
POSTGRES_DB="kir-tube"

DSN="host=localhost user=postgres password=postgres dbname=kir-tube port=5432 sslmode=disable"
TEST_DSN="host=localhost user=postgres password=postgres dbname=kir-tube_test port=5432 sslmode=disable"

MINIO_ROOT_USER=minio
MINIO_ROOT_PASSWORD=minio12345
MINIO_ENDPOINT=localhost:9000
MINIO_BUCKET=kir-tube
MINIO_USE_SSL=false

MODE=development
SECURE_COOKIES=false
```

> **Важно про `DSN` и Docker.** Значение `host=` в `DSN` — это сетевое имя базы.
> При запуске приложения **внутри Docker Compose** используйте `host=db` (имя
> сервиса БД в `docker-compose.yaml`). При локальном запуске (`go run`) с базой,
> проброшенной на хост, — `host=localhost`.

## 2. Запуск через Docker Compose (рекомендуется)

`docker-compose.yaml` поднимает три сервиса: `backend`, `db` (PostgreSQL),
`minio`.

```sh
docker compose up --build
```

- API: `http://localhost:8000/api`
- MinIO API: `http://localhost:9000`, консоль: `http://localhost:9001`

Остановка и очистка:

```sh
docker compose down          # остановить
docker compose down -v       # + удалить тома (данные БД и MinIO)
```

> Всегда останавливайте стек через `docker compose down`, а не `Ctrl+C`/`stop` —
> иначе контейнер с фиксированным `container_name` может занять имя и вызвать
> конфликт при следующем `up`.

## 3. Локальный запуск (без контейнера приложения)

Поднимите PostgreSQL и MinIO (можно только их из compose):

```sh
docker compose up db minio
```

Затем через `Makefile`:

```sh
make migrate   # go run migrations/auto.go  — создать/обновить схему
make seed      # go run cmd/seed/main.go     — наполнить тестовыми данными
make run       # go run cmd/main.go          — запустить API
make build     # go build cmd/main.go        — собрать бинарь
make prod      # собрать и запустить ./main.exe
```

Или напрямую:

```sh
go run ./cmd/main.go
```

## 4. Миграции

Схема создаётся через GORM `AutoMigrate` в `migrations/auto.go`:

```sh
make migrate
# или
go run migrations/auto.go
```

Мигрируются модели: `Video`, `WatchHistory`, `Playlist`, `Channel`, `User`,
`VideoComment`, `VideoLike` (плюс автоматически — join-таблицы many2many:
`video_tags`, `channel_subscribers`, `playlist_videos`). Подробнее —
[database.md](./database.md).

## 5. Сидинг

```sh
make seed
# или
go run ./cmd/seed/main.go
```

Данные для сидинга лежат в `seeder/data/*.json` (каналы, видео).

## 6. Тестовый аккаунт

Из корневого `CLAUDE.md`:

- **Login:** `test@test.ru`
- **Password:** `qwerty123`

## Полезные команды разработки

```sh
go build ./...            # собрать всё
go test ./...             # все тесты
go test -race -cover ./...# гонки + покрытие
gofmt -w .                # форматирование
go vet ./...              # статический анализ
go mod tidy               # привести зависимости в порядок
```

Подробнее о тестах — [testing.md](./testing.md).
