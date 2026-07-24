<h1 align="center">🎬 Kir-tube Backend</h1>

<p align="center">
  Бэкенд видеоплатформы («туб») на Go — модульный монолит с чистой архитектурой,
  JWT-аутентификацией, транскодированием видео и объектным хранилищем.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white" alt="Go 1.26">
  <img src="https://img.shields.io/badge/PostgreSQL-GORM-4169E1?logo=postgresql&logoColor=white" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/MinIO-S3-C72E49?logo=minio&logoColor=white" alt="MinIO">
  <img src="https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white" alt="Docker">
  <img src="https://img.shields.io/badge/JWT-Auth-000000?logo=jsonwebtokens&logoColor=white" alt="JWT">
</p>

---

## 📖 О проекте

**Kir-tube** — серверная часть видеохостинга: пользователи, каналы, видео,
плейлисты, комментарии, лайки, подписки, история просмотров и загрузка медиа с
автоматическим транскодированием в несколько разрешений.

Порт NestJS-бэкенда на стандартную библиотеку Go (`net/http`) с сохранением
формата данных (совместимые id, имена таблиц) — миграция без потери совместимости
с фронтендом.

### Возможности

- 🔐 **Аутентификация** — регистрация, вход, JWT (access + refresh), refresh в HttpOnly-cookie, верификация email
- 📺 **Видео** — каталог, поиск, тренды, персональные рекомендации, счётчик просмотров, лайки
- 📡 **Каналы** — страницы каналов, подписки/отписки
- 🎧 **Плейлисты** — создание, добавление/удаление видео
- 💬 **Комментарии** — CRUD с проверкой владельца
- 🎛 **Студия** — кабинет автора: управление своими видео
- 🖼 **Медиа** — загрузка файлов в MinIO, транскодирование видео через ffmpeg, безопасная раздача статики

---

## 🏛 Архитектурный подход

Проект следует принципам **Clean Architecture**, **SOLID** и **DDD** — зависимости
направлены внутрь, к бизнес-логике; программируем на интерфейсы, а не реализации.

### Модульный монолит

Каждый бизнес-домен — самостоятельный модуль в `internal/`. Все модули собираются
в единый `http.Handler` функцией `app.Assemble`, а `cmd/main.go` лишь оборачивает
его в HTTP-сервер. Благодаря этому тесты собирают всё приложение с подменёнными
зависимостями (тестовая БД, in-memory хранилище).

### Слои

```
   HTTP-запрос
       │
   ┌───▼─────────┐   разбор запроса / сериализация ответа
   │  Handler    │   (роуты на http.ServeMux)
   └───┬─────────┘
   ┌───▼─────────┐   бизнес-правила, оркестрация
   │  Service    │   (зависит от интерфейсов, не реализаций)
   └───┬─────────┘
   ┌───▼─────────┐   доступ к данным (GORM)
   │ Repository  │
   └───┬─────────┘
       ▼
   PostgreSQL / MinIO
```

### Ключевые решения

- **Dependency Injection вручную** через конструкторы — без DI-фреймворка. Точка
  проводки — `internal/app/app.go`.
- **Межконтекстные порты** (`pkg/di`) — узкие интерфейсы разрывают циклы импортов
  между доменами. `pkg` не зависит от `internal`.
- **Единый `/api`-префикс** через `http.StripPrefix`: модули регистрируют «чистые»
  пути, не зная про префикс. Публичная раздача медиа живёт вне `/api` — на `/uploads`.
- **Стабильная JSON-форма**: коллекции сериализуются как `[]`, а не `null`.

### Структура

```
cmd/
  main.go            # точка входа HTTP-сервера
  seed/main.go       # наполнение БД тестовыми данными
configs/             # загрузка конфигурации из окружения
internal/            # приватный код приложения (домены)
  app/               # сборка всех модулей в один handler
  auth/ user/ video/ channel/ playlist/ studio/ comment/ media/ history/
migrations/auto.go   # GORM AutoMigrate
pkg/                 # переиспользуемые библиотеки (db, jwt, middleware, req, res, di, gormx…)
seeder/              # логика и данные для сидинга
docs/                # подробная документация
```

📚 **Подробная документация** — в каталоге [`docs/`](./docs):
[архитектура](./docs/architecture.md) ·
[API](./docs/api.md) ·
[база данных](./docs/database.md) ·
[медиа](./docs/media.md) ·
[конфигурация](./docs/configuration.md) ·
[тесты](./docs/testing.md)

---

## 🧰 Технологический стек

| Категория | Технологии |
|-----------|-----------|
| Язык / HTTP | Go 1.26, стандартный `net/http` (`ServeMux` с method+path паттернами) |
| База данных | PostgreSQL, GORM (`gorm.io/driver/postgres`, драйвер `pgx/v5`) |
| Хранилище | MinIO / S3 (`minio-go/v7`) |
| Аутентификация | JWT HS256 (`golang-jwt/v5`), bcrypt (`golang.org/x/crypto`) |
| Валидация | `go-playground/validator/v10` |
| Конфиг | `joho/godotenv` |
| Медиа | внешние `ffmpeg` / `ffprobe` |
| Тесты | `stretchr/testify` |

---

## 🚀 Быстрый старт

### Через Docker Compose (рекомендуется)

Поднимает `backend` + PostgreSQL + MinIO:

```sh
docker compose up --build
```

- API — `http://localhost:8000/api`
- Публичные медиафайлы — `http://localhost:8000/uploads/...`
- MinIO консоль — `http://localhost:9001`

Остановка: `docker compose down` (с `-v` — удалить и тома с данными).

### Локально

```sh
docker compose up db minio   # поднять зависимости
make migrate                 # создать схему
make seed                    # наполнить тестовыми данными
make run                     # запустить API
```

> Перед запуском создайте `.env` в корне — список переменных в
> [`docs/configuration.md`](./docs/configuration.md).

### Тестовый аккаунт

- **Login:** `test@test.ru`
- **Password:** `qwerty123`

---

## 📜 Скрипты

### Makefile

| Команда | Действие |
|---------|----------|
| `make run` | Запустить API (`go run cmd/main.go`) |
| `make build` | Собрать бинарь (`go build cmd/main.go`) |
| `make prod` | Собрать и запустить `./main.exe` |
| `make migrate` | Применить схему БД (`go run migrations/auto.go`) |
| `make seed` | Наполнить БД тестовыми данными (`go run cmd/seed/main.go`) |

### Разработка

```sh
go build ./...             # собрать всё
go test ./...              # все тесты
go test -short ./...       # только юнит (без БД)
go test -race -cover ./... # гонки + покрытие
gofmt -w .                 # форматирование
go vet ./...               # статический анализ
go mod tidy                # привести зависимости в порядок
```

---

## 🗺 Обзор API

Базовый URL — `/api`. Защищённые эндпоинты требуют `Authorization: Bearer <token>`.

| Группа | Примеры эндпоинтов |
|--------|--------------------|
| **Auth** | `POST /auth/register`, `POST /auth/login`, `POST /auth/access-token`, `POST /auth/logout` |
| **Users** | `GET /users/profile`, `PUT /users/profile` |
| **Videos** | `GET /videos`, `GET /videos/trending`, `GET /videos/explore`, `GET /videos/by-publicId/{id}` |
| **Channels** | `GET /channels`, `GET /channels/by-slug/{slug}`, `PATCH /channels/toggle-subscribe/{slug}` |
| **Playlists** | `GET /playlists`, `POST /playlists`, `POST /playlists/{id}/toggle-video` |
| **Studio** | `GET/POST/PUT/DELETE /studio/videos` |
| **Comments** | `GET /comments/by-video/{id}`, `POST /comments`, `PUT/DELETE /comments/{id}` |
| **Media** | `POST /upload-file`, `GET /upload-file/status/{name}`, `GET /uploads/{path}` |

📘 Полный справочник со всеми телами и кодами ответов — [`docs/api.md`](./docs/api.md).

---

## 📬 Контакты

Автор — **Кирилл Вегеле**

<p>
  <a href="https://t.me/kerrove">
    <img src="https://img.shields.io/badge/Telegram-@kerrove-26A5E4?logo=telegram&logoColor=white" alt="Telegram">
  </a>
  <a href="mailto:kirove.work@gmail.com">
    <img src="https://img.shields.io/badge/Email-kirove.work@gmail.com-EA4335?logo=gmail&logoColor=white" alt="Email">
  </a>
</p>

- 💬 Telegram: [@kerrove](https://t.me/kerrove)
- 📧 Email: [kirove.work@gmail.com](mailto:kirove.work@gmail.com)
