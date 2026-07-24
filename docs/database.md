# Модель данных

БД — PostgreSQL, доступ через GORM. Схема создаётся `AutoMigrate`
(`migrations/auto.go`). Имена таблиц заданы явно через `TableName()` и совпадают
с именами из прежней Prisma-схемы (`@@map`).

## Общие базовые типы (`pkg/gormx`)

Каждая модель встраивает один из базовых типов:

- **`Identifier`** — строковый PK `id varchar(30)`, генерируется в `BeforeCreate`,
  если не задан (cuid-подобный: `c` + 24 base36-символа).
- **`Base`** — `Identifier` + `CreatedAt`.
- **`TimestampedBase`** — `Base` + `UpdatedAt`.

## Таблицы

### `user` — пользователь
| Поле | Тип | Ограничения |
|------|-----|-------------|
| `id` | varchar(30) | PK |
| `name` | text? | nullable |
| `email` | text | **unique**, not null |
| `password` | text | not null, скрыт из JSON (`json:"-"`) |
| `verification_token` | text? | скрыт из JSON; генерируется в `BeforeCreate`, если пуст |
| `created_at`, `updated_at` | timestamptz | |

### `channel` — канал (публичное лицо пользователя)
| Поле | Тип | Ограничения |
|------|-----|-------------|
| `id` | varchar(30) | PK |
| `slug` | text | **unique**, not null |
| `description` | text? | nullable |
| `is_verified` | bool | not null, default `false` |
| `avatar_url`, `banner_url` | text? | nullable |
| `user_id` | varchar(30) | **unique** FK → `user`, cascade delete (один канал на пользователя) |
| `created_at`, `updated_at` | timestamptz | |

Связи: `Subscribers` — many-to-many с `user` через join-таблицу
`channel_subscribers`.

### `video` — видео
| Поле | Тип | Ограничения |
|------|-----|-------------|
| `id` | varchar(30) | PK |
| `public_id` | text | **unique**, not null (короткий публичный id) |
| `title` | text | not null |
| `description` | text | not null |
| `thumbnail_url` | text | not null |
| `video_file_name` | text | not null |
| `max_resolution` | text | not null, default `1080p` |
| `views_count` | int | not null, default `0` |
| `is_public` | bool | not null, default `false` |
| `channel_id` | varchar(30) | index, not null, FK → `channel`, cascade delete |
| `created_at`, `updated_at` | timestamptz | |

Связи: `Comments` (1-N), `Likes` (1-N), `Tags` (M-N через `video_tags`).

### `video_comment` — комментарий
| Поле | Тип | Ограничения |
|------|-----|-------------|
| `id` | varchar(30) | PK |
| `text` | text | not null |
| `user_id` | varchar(30) | index, not null, FK → `user`, cascade |
| `video_id` | varchar(30) | index, not null, FK → `video`, cascade |
| `created_at`, `updated_at` | timestamptz | |

### `video_like` — лайк
| Поле | Тип | Ограничения |
|------|-----|-------------|
| `id` | varchar(30) | PK |
| `user_id` | varchar(30) | not null, FK → `user`, cascade |
| `video_id` | varchar(30) | not null, FK → `video`, cascade |
| `created_at` | timestamptz | (только `Base`, без `updated_at`) |

Композитный **уникальный индекс** `idx_video_like_user_video` по (`user_id`,
`video_id`) — один лайк на пару пользователь/видео.

### `video_tag` — тег
| Поле | Тип | Ограничения |
|------|-----|-------------|
| `id` | varchar(30) | PK |
| `name` | text | **unique**, not null |
| `created_at`, `updated_at` | timestamptz | |

Связь с `video` — M-N через `video_tags`.

### `playlist` — плейлист
| Поле | Тип | Ограничения |
|------|-----|-------------|
| `id` | varchar(30) | PK |
| `title` | text | not null |
| `user_id` | varchar(30) | index, not null, FK → `user`, cascade |
| `created_at`, `updated_at` | timestamptz | |

Связь с `video` — M-N через `playlist_videos`.

### `watch_history` — история просмотров
| Поле | Тип | Ограничения |
|------|-----|-------------|
| `id` | varchar(30) | PK |
| `user_id` | varchar(30) | not null, FK → `user`, cascade |
| `video_id` | varchar(30) | not null, FK → `video`, cascade |
| `watched_at` | timestamptz | autoCreateTime |

Композитный **уникальный индекс** `idx_watch_history_user_video` по (`user_id`,
`video_id`) — одна запись на пару, обновляется при повторном просмотре.

## Join-таблицы (many-to-many)

Создаются GORM автоматически:

| Таблица | Связь |
|---------|-------|
| `channel_subscribers` | `channel` ↔ `user` (подписчики) |
| `video_tags` | `video` ↔ `video_tag` |
| `playlist_videos` | `playlist` ↔ `video` |

## Схема связей

```
user 1───1 channel 1───N video ───N video_comment
 │            (subscribers M─N user)   ├──N video_like
 │                                     └──M─N video_tag  (video_tags)
 ├──N playlist ──M─N video  (playlist_videos)
 ├──N video_comment
 ├──N video_like
 └──N watch_history ──N video
```

## Читаемые модели (не таблицы)

Некоторые типы — проекции для ответов, а не таблицы:

- **`channel.ChannelDetails`** — канал + его владелец, подписчики и видео. Видео
  подтягиваются из домена `video` через порт `di.IChannelVideoRepository`.
- **`channel.ChannelVideo`** — видео канала с вложенным каналом (чтобы клиент
  читал `video.channel.user` без второго запроса).
- **`di.ContextUser` / `di.ContextChannel`** — аутентифицированный пользователь в
  контексте запроса.

## Стабильность JSON-формы

В хуках `AfterFind` коллекции инициализируются пустыми срезами, чтобы в JSON
сериализовались как `[]`, а не `null`, даже если ассоциации не были предзагружены
(`Video.Comments/Likes/Tags`, `Channel.Subscribers`).
