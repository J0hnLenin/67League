# 6-7 Лига

Студенческая сезонная онлайн-лига по игре Го. Дивизионы, клубы, бейджи, титулы и внутриигровая валюта.

## Что это

**6-7 Лига** — самостоятельный проект для студентов, преподавателей и выпускников вузов, которые играют в Го. Название — отсылка к мему «6-7».

Формат: сезонная лига с дивизионами, клубами и геймификацией. Играть можно на OGS, KGS, Fox или оффлайн.

## Как это работает

- **Сезон** — 1 месяц. Статусы: announced → registration → active → closing → finished → archived.
- **Дивизионы** — S, A, B. Повышение и понижение по итогам сезона (k1, k2, k3).
- **Партии** — игроки сами создают вызовы. Жизненный цикл: proposed → accepted → playing → pending_confirmation → confirmed.
- **Клубы** — объединение по городу, вузу или интересам. Клубный зачёт — сумма LP активных участников.
- **Бейджи** — личные, сезонные, клубные, супер. Дают титулы.
- **Супертитул «Чемпион 67»** — лучший игрок лучшего клуба сезона.

## Валюты

- **LP** — соревновательный рейтинг. Не тратится.
- **Камни** — валюта для косметики. Не даёт игрового преимущества.

## Стек

| Компонент | Технология |
|-----------|-----------|
| Frontend | Next.js 15 + React 19 + TypeScript + Tailwind CSS v4 |
| Backend | Go 1.27 + Chi + pgx |
| База данных | PostgreSQL 16 + goose миграции |
| Auth | JWT + bcrypt |
| Email | SMTP / LogSender (dev) |
| Тесты | testify + mockery + testcontainers |

## Структура репозитория

```
67League/
├── src/
│   ├── backend/          # Go бекенд
│   │   ├── cmd/api/      # API-сервер (:8080)
│   │   ├── cmd/admin/    # Admin-сервер (:8081)
│   │   ├── internal/
│   │   │   ├── domain/   # Доменные сущности
│   │   │   ├── application/  # Use cases + порты
│   │   │   ├── infrastructure/  # Postgres, auth, email
│   │   │   └── delivery/   # HTTP handlers
│   │   └── migrations/   # SQL миграции
│   └── frontend/         # Next.js фронтенд
│       ├── app/          # App Router страницы
│       └── lib/          # API клиент, типы
├── wiki/                 # Документация
│   ├── task.md           # Полное ТЗ
│   ├── DEVELOPMENT_MANIFEST.md
│   ├── business-logic.md
│   └── badges.md
└── .github/workflows/    # CI/CD
```

## Быстрый старт

### Весь стек одной командой (Docker Compose)

```bash
cd /home/sente/projects/67League

# Скопировать env-файл (опционально)
cp .env.example .env

# Поднять всё: Postgres + миграции + бекенд + админ + фронтенд
docker compose up --build

# Открыть:
#   Фронтенд:   http://localhost:3000
#   API:        http://localhost:8080/api/v1
#   Admin API:  http://localhost:8081
```

### Остановить

```bash
docker compose down

# Удалить данные БД тоже:
docker compose down -v
```

### Локальная разработка (без Docker)

**Бекенд:**
```bash
cd src/backend
docker compose up -d postgres   # только БД
make migrate-up                 # применить миграции
go run cmd/api/main.go          # API на :8080
go run cmd/admin/main.go        # Admin на :8081
make test                       # тесты
```

**Фронтенд:**
```bash
cd src/frontend
npm install
npm run dev        # dev-сервер на :3000
npm run build      # production билд
```

## Статус (2026-10-19)

### ✅ Сделано

- **Backend**: Clean Architecture, 17 таблиц БД, миграции
- **Domain**: Player, Match, Season, Division, Club, Badge, Settings — модели + тесты
- **Use cases**: регистрация, вход, клубы, сезоны, партии (полный жизненный цикл), бейджи, ротация
- **HTTP API**: auth, player, club, season, division, match (публичные + приватные)
- **Admin API**: auth, season management, player management, settings
- **Frontend**: Next.js + Tailwind, главная, авторизация, регистрация, dashboard, клубы, сезоны
- **CI/CD**: GitHub Actions workflow (backend + frontend)
- **Документация**: ТЗ, манифест, бизнес-логика, бейджи, QA

### 🚧 В процессе / TODO

- Дописать страницы фронтенда (партии, профиль, админка)
- Интеграционные тесты (требуют Docker)
- Email-уведомления (заглушка → SMTP)
- OGS API интеграция
- Магазин и покупки
- S3-хранилище для SGF
- E2E тесты (Playwright)

## Документы

- `wiki/task.md` — полное техническое задание
- `wiki/DEVELOPMENT_MANIFEST.md` — архитектура, процесс, quality gates
- `wiki/business-logic.md` — правила лиги, формулы
- `wiki/badges.md` — список бейджей и титулов
- `wiki/QA.md` — ответы на ключевые вопросы

## Лицензия

MIT
