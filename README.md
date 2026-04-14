# URL Shortener

Сервис сокращения ссылок с веб-интерфейсом и аналитикой посещений.

### Hexlet tests and linter status:
[![Actions Status](https://github.com/MakhovRoman/go-project-278/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/MakhovRoman/go-project-278/actions) [![CI](https://github.com/MakhovRoman/go-project-278/actions/workflows/workflow.yml/badge.svg)](https://github.com/MakhovRoman/go-project-278/actions/workflows/workflow.yml) [![Render](https://img.shields.io/badge/Render-deployed-brightgreen?logo=render)](https://go-project-278-15gh.onrender.com)

## Стек

- **Go** + **Gin** — HTTP-сервер
- **PostgreSQL** — база данных
- **sqlc** — генерация Go-кода из SQL-запросов
- **goose** — миграции базы данных
- **Caddy** — reverse proxy, раздача статики
- **GlitchTip** — сбор ошибок (Sentry-совместимый)
- **Render** — деплой

## Возможности

- Создание коротких ссылок с автогенерацией или кастомным `short_name`
- Редирект по короткой ссылке с записью статистики посещений
- CRUD API для управления ссылками
- Пагинация списков через параметр `range=[offset,limit]`
- Валидация входящих данных
- Веб-интерфейс

## API

### Ссылки

| Метод | Маршрут | Описание |
|-------|---------|----------|
| `GET` | `/api/links?range=[0,10]` | Список ссылок с пагинацией |
| `POST` | `/api/links` | Создать ссылку |
| `GET` | `/api/links/:id` | Получить ссылку по ID |
| `PUT` | `/api/links/:id` | Обновить ссылку |
| `DELETE` | `/api/links/:id` | Удалить ссылку |
| `GET` | `/r/:code` | Редирект по короткому имени |

### Аналитика

| Метод | Маршрут | Описание |
|-------|---------|----------|
| `GET` | `/api/link_visits?range=[0,10]` | Список посещений с пагинацией |

### Пример создания ссылки

```bash
curl -X POST https://go-project-278-15gh.onrender.com/api/links \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://example.com","short_name":"example"}'
```

Ответ:
```json
{
  "id": 1,
  "original_url": "https://example.com",
  "short_name": "example",
  "short_url": "https://go-project-278-15gh.onrender.com/r/example",
  "created_at": "2025-01-01T00:00:00Z"
}
```

## Локальная разработка

### Требования

- Go 1.25+
- Node.js 20+
- PostgreSQL
- [goose](https://github.com/pressly/goose)
- [sqlc](https://sqlc.dev)

### Установка

```bash
git clone https://github.com/MakhovRoman/go-project-278
cd go-project-278
go mod download
npm install
```

### Настройка окружения

Создайте файл `.env`:

```env
DATABASE_URL=postgres://postgres:password@localhost:5432/appdb?sslmode=disable
BASE_URL=http://localhost:8080
PORT=8080
SENTRY_DSN=           # опционально
```

### Миграции

```bash
goose -dir ./db/migrations postgres "$DATABASE_URL" up
```

### Запуск

```bash
# Бэкенд + фронтенд вместе
npm start

# Только бэкенд
go run .
```

Фронтенд доступен на http://localhost:5173, API — на http://localhost:8080.

### Тесты

```bash
make test
```

### Регенерация DB-кода

```bash
sqlc generate
```
