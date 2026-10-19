Часть I. Манифест разработки «6-7 Лига»
0. Миссия

«6-7 Лига» — студенческая сезонная онлайн-лига по игре Го.
Цель: дать сообществу регулярную соревновательную площадку с прозрачной ротацией, клубами, бейджами, титулами и внутриигровой валютой, не превращая проект в pay-to-win.
1. Инженерные ценности

    Бизнес-логика важнее кода. Сначала правила, инварианты, состояния и сценарии. Только потом модели, интерфейсы и реализация.

    Чистая архитектура. Зависимости направлены внутрь. Домен ничего не знает о БД, HTTP, Next.js, SMTP и S3.

    Go как основной язык бекенда. Простота, явность, тестируемость, минимальная магия.

    Контракты раньше реализации. Интерфейсы и тесты фиксируют поведение до написания имплементации.

    Все числовые константы — настраиваемые. k1, k2, k3, p, n, t, сроки, бонусы, LP и камни хранятся в БД и меняются через админку.

    Один источник правды. Правила лиги живут в домене, а не в контроллерах, SQL или фронтенде.

    Автоматизация вместо ручной работы. Бейджи, LP, камни, ротации, стрики, клубный зачёт считаются системой.

    Расширяемость. Сейчас 3 дивизиона S/A/B, но архитектура не должна ломаться при 2, 5 или 10 дивизионах.

    Безопасность и приватность. Пароли хешируются, email не публикуется, админ — отдельная сущность и отдельный вход.

    Документируемость. ADR, OpenAPI, схема БД, глоссарий и правила лежат в репозитории.

2. Архитектура

Используем Clean Architecture / Hexagonal Architecture.
Слои
Слой	Содержание	Зависимости
Domain	Сущности, value objects, доменные правила, ошибки, события	Ни от чего не зависит
Application	Use cases, ports, DTO, транзакционные границы	Зависит только от domain
Infrastructure	Postgres, S3, email, OGS API, auth, config	Реализует ports
Delivery	HTTP API, admin API, Next.js frontend, CLI	Зависит от application
Правило зависимостей
text

domain <- application <- infrastructure/delivery

Нельзя импортировать infrastructure в domain или application.
Нельзя использовать ORM-модели как доменные сущности.
Интерфейсы объявляются там, где они нужны потребителю: чаще всего в application/ports.
Пример структуры Go-проекта
text

cmd/
  api/
  admin/
internal/
  domain/
    player/
    club/
    season/
    division/
    match/
    badge/
    settings/
  application/
    ports/
    usecases/
    dto/
  infrastructure/
    persistence/postgres/
    email/
    storage/s3/
    auth/
    config/
  delivery/
    http/
    adminhttp/
migrations/
web/
e2e/

3. Процесс разработки

Порядок строго фиксирован.
Этап 0. Discovery

    Собрать глоссарий.

    Зафиксировать открытые вопросы.

    Описать жизненные циклы: игрок, сезон, партия.

    Составить список настроек сезона и сайта.

    Завести ADR на ключевые решения.

Этап 1. Бизнес-логика

Артефакты:

    Use cases.

    Инварианты.

    Таблицы состояний и переходов.

    Формулы LP, камней, стриков, клубного зачёта, ротации.

    Список событий и уведомлений.

Без этого этапа код не пишется.
Этап 2. Модели и интерфейсы

    Доменные сущности и value objects.

    Порты репозиториев и сервисов.

    DTO для API.

    Доменные ошибки.

    События домена.

Этап 3. Табличные тесты под интерфейсы

    Go table-driven tests.

    Тесты фиксируют контракт до имплементации.

    Используем fakes/mocks.

    Покрываем успешные сценарии, ошибки, границы, конкурентные случаи.

Этап 4. Имплементация бекенда

    Use cases.

    Доменные сервисы.

    Адаптеры: Postgres, email, S3, auth.

    HTTP handlers.

Этап 5. База данных и фронтенд

    Миграции Postgres.

    sqlc/pgx или аналог.

    Next.js + TypeScript.

    Интеграция с API.

    Админ-панель на отдельном эндпоинте.

Этап 6. E2E-тесты

    Playwright/Cypress для UI.

    Go integration/e2e для API.

    Реальная тестовая БД, миграции, сиды.

    Критические пользовательские пути.

4. Тестирование
Уровень	Что проверяем	Инструменты
Unit	Домен, формулы, состояния	Go testing, table-driven
Use case	Сценарии приложения	Go testing, fakes
Integration	Репозитории, БД, миграции	testcontainers, Postgres
Contract	API, OpenAPI	schemathesis, Go tests
E2E	UI + API + БД	Playwright, Go e2e

Целевое покрытие:

    domain/application — не ниже 90%.

    infrastructure — не ниже 70%.

    e2e — все критические сценарии.

Текущий прогресс (2026-10-19):

    domain — тесты написаны для: player, match, season, division, club, shared, settings.
      Остались: admin, badge, economy, notification.

    application — use case тесты проходят (register, login, club create/join,
      season join, match lifecycle, badge awarding, rotation, calculation).

    infrastructure — тесты написаны, требуют Docker + Postgres.

    e2e — Go e2e для API написаны, требуют Docker + Postgres.
      Playwright/Cypress для UI — не начаты (нет frontend).

5. Quality Gates

Перед merge:

    gofmt, go vet, staticcheck, golangci-lint.

    go test ./... -race.

    Миграции применяются на чистой БД.

    OpenAPI актуален.

    Frontend: lint, typecheck, build.

    E2E на критические сценарии.

6. Безопасность

    Пароли: bcrypt/argon2.

    Админ: отдельная таблица, отдельный эндпоинт, отдельный вход.

    Роли: игрок, админ.

    HTTPS, CSRF, rate limiting.

    AuditLog на все админские действия.

    Email не публикуется.

    SGF и изображения — в S3-совместимом хранилище.

7. Definition of Done

Фича готова, если:

    Описана бизнес-логика.

    Есть модели и интерфейсы.

    Написаны табличные тесты.

    Есть имплементация.

    Есть миграции.

    Есть API и UI.

    Есть e2e-сценарий.

    Обновлены docs/OpenAPI/ADR.

    Пройдены quality gates.