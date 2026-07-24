# Архитектура

## Обзор

Kir-tube — **модульный монолит**. Каждый бизнес-домен живёт в отдельном пакете
`internal/<domain>` и оформлен как **модуль** с единой точкой сборки
(`New<Domain>Module`). Все модули собираются в один `http.Handler` функцией
`app.Assemble`, а `cmd/main.go` лишь оборачивает его в `http.Server`.

Такое разделение (`cmd/main.go` → `internal/app` → модули) позволяет тестам
собирать всё приложение с подменёнными зависимостями (тестовая БД, in-memory
хранилище) вместо продакшн-реализаций.

## Слои (Clean Architecture)

Зависимости направлены **внутрь**, к бизнес-логике. Внутри каждого модуля:

```
Handler  →  Service  →  Repository  →  БД / внешние системы
(HTTP)      (бизнес-   (доступ к
            логика)     данным)
```

- **Handler** (`handler.go`) — HTTP-адаптер: разбирает запрос, вызывает сервис,
  сериализует ответ. Регистрирует роуты на `http.ServeMux`.
- **Service** (`service.go`) — бизнес-правила, оркестрация. Зависит от
  **интерфейсов** репозиториев, не от конкретных реализаций.
- **Repository** (`repository.go`) — доступ к данным через GORM.
- **Model** (`model.go`) — доменные сущности (GORM-модели).
- **Payload** (`payload.go`) — DTO запросов/ответов с тегами `validate`.
- **Module** (`module.go`) — композиция: создаёт репозиторий → сервис → хендлер и
  внедряет зависимости (Constructor Injection).

## Dependency Injection

DI выполняется **вручную** через конструкторы — без DI-фреймворка. Точка проводки —
`internal/app/app.go`:

```go
func Assemble(conf *configs.Config, database *db.Db, storage media.ObjectStorage) (http.Handler, string) {
    router := http.NewServeMux()
    apiRouter := http.NewServeMux()

    // Общие зависимости, разделяемые между доменами:
    videoRepository := video.NewVideoRepository(database)
    userProvider := channel.NewContextUserRepository(database)

    userModule := user.NewUserModule(router, user.UserModuleDeps{ ... })
    auth.NewAuthModule(router, auth.AuthModuleDeps{ UserService: userModule.UserService, ... })
    video.NewVideoModule(router, ...)
    channel.NewChannelModule(router, ...)
    playlist.NewPlaylistModule(router, ...)
    studio.NewStudioModule(router, ...)
    comment.NewCommentModule(router, ...)
    media.NewMediaModule(router, ...)

    // /api-префикс + middleware (CORS, логирование)
    apiRouter.Handle("/api/", http.StripPrefix("/api", router))
    stack := middleware.Chain(middleware.CORS(conf.Network.ClientUrl), middleware.Logging)
    return stack(apiRouter), conf.Network.Port
}
```

Каждый модуль принимает свою структуру `...ModuleDeps` — явный контракт
зависимостей.

## Маршрутизация и префиксы

Два вложенных мультиплексора:

- **`router`** (внутренний) — на нём модули регистрируют пути вида
  `POST /comments`, `GET /videos/...`.
- **`apiRouter`** (корневой) — монтирует `router` под префиксом `/api` через
  `http.StripPrefix`. Поэтому модули не знают про `/api` — они регистрируют
  «чистые» пути, а префикс снимается до них.

Итог:

- Все API-эндпоинты доступны под **`/api/...`**.
- **Медиа-модуль** — исключение: статические файлы отдаются на корневом уровне по
  `/uploads/...` (вне `/api`), а также с обратно-совместимым алиасом
  `/api/uploads/...`. Загрузка/статус (`/upload-file`) — обычные API-роуты под
  `/api`. См. [media.md](./media.md).

Маршруты регистрируются через `logs.RouteLog`, который оборачивает хендлер в
логирование и печатает таблицу зарегистрированных роутов при старте.

## Межконтекстные порты (пакет `pkg/di`)

Домены не импортируют друг друга напрямую там, где это создало бы **цикл
импортов**. Вместо этого в `pkg/di` объявлены узкие интерфейсы-порты (принцип
Interface Segregation + Dependency Inversion). `pkg` не зависит от `internal`.

Ключевые порты:

- **`di.IUserProvider`** — `FindContextUser(id) (*ContextUser, error)`. Нужен
  middleware аутентификации, чтобы по id из токена загрузить пользователя вместе
  с его каналом. Реализуется `channel.ContextUserRepository` (он читает и таблицу
  пользователей, и каналов — чего сам пакет `user` не может без цикла).
- **`di.ContextUser` / `di.ContextChannel`** — аутентифицированный пользователь,
  которого middleware кладёт в контекст запроса. Пароль и токен верификации
  скрыты из JSON.
- **`di.IVideoRepository`, `di.IChannelVideoRepository`, `di.IPlaylistVideoRepository`**
  — узкие контракты для чтения видео из доменов `user`, `channel`, `playlist` без
  импорта пакета `video`.

Единственная конкретная реализация видео-репозитория — `video.VideoRepository` —
удовлетворяет всем этим портам сразу (компайл-тайм проверки в
`internal/video/ports.go`):

```go
var (
    _ IVideoRepository            = (*VideoRepository)(nil)
    _ di.IVideoRepository         = (*VideoRepository)(nil)
    _ di.IChannelVideoRepository  = (*VideoRepository)(nil)
    _ di.IPlaylistVideoRepository = (*VideoRepository)(nil)
)
```

## Модули и их назначение

| Модуль | Ответственность |
|--------|-----------------|
| `auth` | Регистрация, вход, выход, refresh-токены, верификация email |
| `user` | Профиль текущего пользователя (чтение/обновление) |
| `video` | Публичный каталог видео: списки, поиск, тренды, рекомендации, просмотры, лайки |
| `channel` | Каналы: список, страница по slug, подписка/отписка |
| `playlist` | Плейлисты пользователя: создание, просмотр, добавление/удаление видео |
| `studio` | Кабинет автора: CRUD своих видео (в рамках своего канала) |
| `comment` | Комментарии к видео: чтение, создание, редактирование, удаление |
| `media` | Загрузка файлов, транскодирование видео, раздача статики из MinIO |
| `history` | Модель истории просмотров (`watch_history`) |

## Middleware

Стек собирается через `middleware.Chain` и применяется ко всему `apiRouter`:

- **`CORS(clientURL)`** — разрешает запросы только с origin, равного `CLIENT_URL`;
  отдаёт `Access-Control-Allow-Credentials: true`, отвечает на preflight `OPTIONS`
  статусом 204.
- **`Logging`** — логирование запросов.

Точечные (per-route) middleware аутентификации:

- **`IsAuthed`** — требует валидный `Authorization: Bearer <accessToken>`; грузит
  `ContextUser` в контекст. Без токена — 401.
- **`MaybeAuthed`** — токен опционален: если валиден, кладёт пользователя в
  контекст, иначе пропускает дальше как анонима (используется в
  `GET /videos/explore` для персонализации).

## Общие пакеты (`pkg/`)

| Пакет | Назначение |
|-------|-----------|
| `pkg/db` | Инициализация GORM-соединения (`NewDb`) |
| `pkg/gormx` | Базовые модели: `Identifier` (cuid-подобный строковый PK), `Base`, `TimestampedBase`; генерация ID |
| `pkg/jwt` | Создание и парсинг JWT (HS256) |
| `pkg/password` | Хеширование и проверка паролей (bcrypt) |
| `pkg/middleware` | CORS, логирование, аутентификация, `Chain` |
| `pkg/req` | Декодирование тела запроса, валидация, извлечение id профиля из контекста |
| `pkg/res` | Запись JSON-ответов и ошибок |
| `pkg/di` | Межконтекстные порты и DTO контекста |
| `pkg/logs` | Логирование маршрутов (`RouteLog`) |
| `pkg/testutil` | Хелперы тестов: тестовая БД, фикстуры, фейковое хранилище |

## Идентификаторы

Две схемы ID (совместимость с данными из прежнего NestJS-бэкенда):

- **Первичные ключи** — cuid-подобные строки: префикс `c` + 24 base36-символа
  (`gormx.Identifier`, `varchar(30)`). Генерируются в GORM-хуке `BeforeCreate`,
  если не заданы.
- **Публичные id видео** — короткие URL-safe строки из 10 base36-символов
  (`video.NewPublicID`, аналог `nanoid(10)`). Отделены от PK: во внешних URL
  фигурирует именно публичный id, а PK остаётся внутренним.
