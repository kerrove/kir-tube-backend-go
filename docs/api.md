# Справочник HTTP API

- **Базовый URL:** `http://localhost:8000/api`
- **Формат:** JSON (`Content-Type: application/json`)
- **Публичные медиафайлы:** `http://localhost:8000/uploads/...` (вне `/api`;
  есть обратно-совместимый алиас `/api/uploads/...`)

## Аутентификация

Защищённые эндпоинты требуют заголовок:

```
Authorization: Bearer <accessToken>
```

- **Access-токен** живёт **1 час**, **refresh-токен** — **720 часов (30 дней)**
  (`internal/auth/service.go`). Оба — JWT HS256 с полями `id`, `isAdmin`, `exp`.
- При `login`/`register`/`access-token` refresh-токен также кладётся в
  **HttpOnly-cookie** `refreshToken` (`SameSite=Lax`, `Secure` = `SECURE_COOKIES`).
- Обновление пары токенов (`POST /auth/access-token`) читает refresh-токен
  **из cookie**.

Обозначения ниже: 🔓 — публичный, 🔒 — требует Bearer (`IsAuthed`),
🔓/🔒 — опциональная авторизация (`MaybeAuthed`).

---

## Auth — `/auth`

### 🔓 POST `/auth/register`
Регистрация. Тело:
```json
{ "email": "user@example.com", "password": "secret" }
```
Валидация: `email` — required+email, `password` — required.
Ответ `200`: объект `AuthUser` (`{ user, accessToken, refreshToken }`), плюс
cookie `refreshToken`. При существующем email — `401`.

### 🔓 POST `/auth/login`
Вход. Тело — как у register. Ответ `200`: `AuthUser` + cookie. Неверные
данные — `401`.

### 🔓 POST `/auth/access-token`
Обновление токенов по refresh-cookie. Без cookie или при невалидном токене —
`401` и очистка cookie. Ответ `200`: `AuthUser` + новая cookie.

### 🔓 POST `/auth/logout`
Выход. Стирает refresh-cookie. Ответ `200`: `true`.

### 🔓 POST `/verify-email?token=<token>`
> Зарегистрирован по пути `/verify-email` (под `/api` → `POST /api/verify-email`),
> **не** `/auth/verify-email`.

Верификация email по токену из query. Без `token` — `401`. Ответ `200`: `true`.

**Объект `AuthUser`:**
```json
{
  "user": { "id": "...", "email": "...", "name": null, "createdAt": "...", "updatedAt": "..." },
  "accessToken": "<jwt>",
  "refreshToken": "<jwt>"
}
```
> Поля `password` и `verificationToken` из `user` никогда не сериализуются.

---

## User — `/users`

### 🔒 GET `/users/profile`
Профиль текущего пользователя (с его каналом). `404`, если пользователь не найден.

### 🔒 PUT `/users/profile`
Обновление профиля. Тело:
```json
{ "name": "New name", "password": "newsecret" }
```
Оба поля опциональны; `password` при наличии — минимум 6 символов. Ответ — профиль.

---

## Videos — `/videos`

### 🔓 GET `/videos?searchTerm=&page=1&limit=10`
Список публичных видео с поиском и пагинацией.

### 🔓 GET `/videos/trending`
Трендовые видео.

### 🔓 GET `/videos/games`
Игровые видео (в текущей реализации отдаёт те же трендовые).

### 🔓/🔒 GET `/videos/explore?page=1&limit=10&excludeIds=id1,id2`
Рекомендации. При наличии валидного токена — персонализированные под пользователя;
`excludeIds` — исключить перечисленные видео.

### 🔓 GET `/videos/by-publicId/{publicId}`
Одно видео по публичному id. `404`, если не найдено.

### 🔓 GET `/videos/by-channel/{channelId}?page=1&limit=10`
Видео канала с пагинацией.

### 🔓 PUT `/videos/update-views-count/{publicId}`
Инкремент счётчика просмотров. Ответ — обновлённое видео.

### 🔒 POST `/users/profile/likes`
Поставить/снять лайк. Тело:
```json
{ "videoId": "<id>" }
```
Ответ:
```json
{ "liked": true }
```

**Пагинация** (везде, где применимо): `page` (по умолчанию 1), `limit`
(по умолчанию 10). Некорректные значения заменяются дефолтами.

---

## Channels — `/channels`

### 🔓 GET `/channels`
Список всех каналов.

### 🔓 GET `/channels/by-slug/{slug}`
Канал по slug вместе с его видео (`ChannelDetails`). `404`, если не найден.

### 🔒 PATCH `/channels/toggle-subscribe/{slug}`
Подписаться/отписаться. Ответ:
```json
{ "message": "Subscribed successfully", "isSubscribed": true }
```

---

## Playlists — `/playlists`  (все 🔒)

### 🔒 GET `/playlists`
Плейлисты текущего пользователя.

### 🔒 GET `/playlists/{playlistId}`
Плейлист по id.

### 🔒 POST `/playlists`
Создать плейлист. Тело:
```json
{ "title": "My playlist", "videoPublicId": "<optional>" }
```
`title` — required; `videoPublicId` — опционален (первое видео).

### 🔒 POST `/playlists/{playlistId}/toggle-video`
Добавить/убрать видео из плейлиста. Тело:
```json
{ "videoId": "<id>" }
```

---

## Studio — `/studio/videos`  (все 🔒, требуется наличие канала)

Кабинет автора. Действия выполняются в рамках канала текущего пользователя;
если у пользователя нет канала — `403`.

### 🔒 GET `/studio/videos?searchTerm=&page=1&limit=10`
Видео своего канала (включая приватные).

### 🔒 GET `/studio/videos/{id}`
Одно своё видео по id.

### 🔒 POST `/studio/videos`
Создать видео. Тело (`CreateVideoInput`):
```json
{
  "title": "Title",
  "description": "",
  "thumbnailUrl": "",
  "videoFileName": "",
  "maxResolution": "1080p",
  "tags": ["go", "tutorial"]
}
```
`title` — required. Публичный id, видимость и привязка к каналу задаются на
сервере. Теги матчатся-или-создаются по имени.

### 🔒 PUT `/studio/videos/{id}`
Частичное обновление (`UpdateVideoInput`) — все поля опциональны (указатели):
```json
{ "title": "New", "isPublic": true, "tags": ["go"] }
```
`nil`-поле не трогается; `tags`, если переданы, **полностью заменяют** набор тегов.

### 🔒 DELETE `/studio/videos/{id}`
Удалить видео.

Ошибки studio: не найдено — `404`, прочее — `500`.

---

## Comments — `/comments`

### 🔓 GET `/comments/by-video/{publicId}`
Комментарии к видео по его публичному id.

### 🔒 POST `/comments`
Создать комментарий. Тело:
```json
{ "text": "Nice!", "videoId": "<id>" }
```
Оба поля — required.

### 🔒 PUT `/comments/{publicId}`
Редактировать свой комментарий. Тело:
```json
{ "text": "Edited" }
```

### 🔒 DELETE `/comments/{publicId}`
Удалить свой комментарий.

Ошибки comments: чужой комментарий — `403`, не найден — `404`, прочее — `500`.

---

## Media — загрузка и раздача

Подробности конвейера — [media.md](./media.md).

### 🔒 POST `/upload-file?folder=<folder>`
Загрузка файла (multipart/form-data, поле формы — `file`).
- Допустимые папки: `avatars`, `banners`, `thumbnails`, `videos` (пусто → `default`).
- Допустимые типы: `image/jpeg`, `image/png`, `image/svg+xml`, `video/mp4`,
  `video/quicktime`.
- Лимит: **300 МБ** на файл.

Ответ `200` — массив дескрипторов:
```json
[{ "url": "/uploads/videos/abc123.mp4", "name": "abc123.mp4", "maxResolution": "1080p" }]
```
`maxResolution` присутствует только для видео. Ошибки валидации — `400`.

### 🔒 GET `/upload-file/status/{fileName}`
Статус транскодирования видео:
```json
{ "fileName": "abc123.mp4", "status": 100 }
```
`status`: `0` — в очереди/неизвестно, `1..100` — прогресс/готово, `-1` — ошибка.

### 🔓 GET `/uploads/{path...}`
Раздача файла из MinIO (например `/uploads/videos/abc123.mp4`). Доступно на
корне и по алиасу `/api/uploads/...`. Обход каталога (`..`) запрещён — `403`.
Листинг каталога (`/uploads`, `/uploads/index.html`) — `403`. Нет файла — `404`.

---

## Коды ошибок

| Код | Когда |
|-----|-------|
| `200` | Успех (в т.ч. на мутациях) |
| `400` | Невалидное тело/параметры, ошибки валидации, ошибки загрузки |
| `401` | Нет/невалидный токен, неверные учётные данные |
| `403` | Доступ запрещён (чужой ресурс, нет канала, обход каталога) |
| `404` | Ресурс не найден |
| `500` | Внутренняя ошибка |

Тело ответа при успехе — JSON (объект, массив или булево). Ошибки отдаются как
текст (`http.Error`) или JSON-строка с сообщением валидатора.
