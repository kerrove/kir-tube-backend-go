# Kir-tube Backend — Документация

Бэкенд видеоплатформы **kir-tube** («туб») на Go. Это порт NestJS-бэкенда на
стандартную библиотеку `net/http` с GORM/PostgreSQL и MinIO для хранения медиа.

Модульный монолит: каждый бизнес-домен — самостоятельный модуль в `internal/`,
все модули собираются в единый `http.Handler` в `internal/app`. Архитектура
следует принципам из корневого `CLAUDE.md`: Dependency Injection, SOLID,
Clean Architecture (зависимости направлены внутрь, к бизнес-логике).

## Оглавление

| Документ | О чём |
|----------|-------|
| [getting-started.md](./getting-started.md) | Установка, запуск локально и в Docker, миграции, сидинг |
| [architecture.md](./architecture.md) | Слои, модули, DI-проводка, порты между контекстами |
| [configuration.md](./configuration.md) | Переменные окружения (`.env`) |
| [api.md](./api.md) | Полный справочник HTTP API |
| [database.md](./database.md) | Модели данных, таблицы, связи |
| [media.md](./media.md) | Загрузка файлов и конвейер транскодирования видео |
| [testing.md](./testing.md) | Юнит-, интеграционные и e2e тесты |

## Технологический стек

- **Язык:** Go 1.26.3, модуль `go/kir-tube`
- **HTTP:** стандартный `net/http` (`http.ServeMux` с method+path-паттернами Go 1.22+)
- **БД:** PostgreSQL через GORM (`gorm.io/driver/postgres`, драйвер `pgx/v5`)
- **Объектное хранилище:** MinIO / S3 (`minio-go/v7`)
- **Аутентификация:** JWT (`golang-jwt/v5`), bcrypt (`golang.org/x/crypto`)
- **Валидация:** `go-playground/validator/v10`
- **Конфиг:** `joho/godotenv`
- **Транскодирование:** внешние `ffmpeg` / `ffprobe`
- **Тесты:** `stretchr/testify`

## Быстрый старт

```sh
cp .env.example .env      # заполнить значения (файла-примера пока нет — см. configuration.md)
docker compose up --build # backend + PostgreSQL + MinIO
```

API доступно на `http://localhost:8000/api`, публичные медиафайлы — на
`http://localhost:8000/uploads/...`.

## Структура репозитория

```
cmd/
  main.go            # точка входа HTTP-сервера
  seed/main.go       # наполнение БД тестовыми данными
configs/             # загрузка конфигурации из окружения
internal/            # приватный код приложения (домены)
  app/               # сборка всех модулей в один handler
  auth/ user/ video/ channel/ playlist/ studio/ comment/ media/ history/
migrations/auto.go   # GORM AutoMigrate
pkg/                 # переиспользуемые библиотеки (db, jwt, middleware, req, res, di, gormx, ...)
seeder/              # логика и данные для сидинга
media/               # исходный NestJS-код (референс для порта, не компилируется)
docs/                # эта документация
```

> Каталог `media/*.ts` — исходный TypeScript/NestJS, оставленный как референс.
> В сборку Go он не входит.
