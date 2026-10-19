# Admin API

Админ API работает на отдельном порту (`:8081`) и требует отдельной аутентификации.

## Аутентификация

Все приватные endpoints требуют заголовок:
```
Authorization: Bearer <admin_token>
```

Токен получается через `POST /admin/login`.

## Endpoints

### Auth

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/admin/login` | — | Вход админа. Body: `{ "login": "...", "password": "..." }`. Returns: `{ "token": "...", "admin_id": N }` |

### Seasons

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/admin/seasons` | ✅ | Список всех сезонов |
| `POST` | `/admin/seasons` | ✅ | Создать сезон. Body: `{ "number": N, "title": "...", "start_date": "RFC3339", "end_date": "RFC3339", "settings": { "k1": N, "k2": N, ... } }` |
| `POST` | `/admin/seasons/:id/start-registration` | ✅ | Перевести в статус `registration` |
| `POST` | `/admin/seasons/:id/start` | ✅ | Перевести в статус `active` |
| `POST` | `/admin/seasons/:id/close` | ✅ | Перевести в статус `closing` |
| `POST` | `/admin/seasons/:id/finish` | ✅ | Перевести в статус `finished` |
| `POST` | `/admin/seasons/:id/archive` | ✅ | Перевести в статус `archived` |
| `POST` | `/admin/seasons/:id/calculate-results` | ✅ | Рассчитать итоги сезона |
| `POST` | `/admin/seasons/:id/apply-rotation` | ✅ | Применить ротацию дивизионов. Body: `{ "next_season_id": N }` |
| `POST` | `/admin/seasons/:id/award-season-badges` | ✅ | Выдать сезонные бейджи |
| `POST` | `/admin/seasons/:id/award-super-title` | ✅ | Выдать супертитул |

### Players

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/admin/players` | ✅ | Список игроков. Query: `?status=active\|inactive\|banned` |
| `GET` | `/admin/players/:id` | ✅ | Детали игрока |
| `PATCH` | `/admin/players/:id` | ✅ | Обновить игрока. Body: `{ "full_name"?, "city"?, "rating"?, "status"?, "club_id"? }` |
| `GET` | `/admin/players/:id/matches` | ✅ | Партии игрока |
| `POST` | `/admin/players/:id/ban` | ✅ | Забанить игрока |
| `POST` | `/admin/players/:id/unban` | ✅ | Разбанить игрока |

### Clubs

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/admin/clubs` | ✅ | Список клубов |

### Matches

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/admin/matches` | ✅ | Список партий |
| `GET` | `/admin/matches/:id` | ✅ | Детали партии |
| `POST` | `/admin/matches/:id/cancel` | ✅ | Отменить партию. Body: `{ "reason"? }` |

### Settings

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/admin/settings` | ✅ | Получить настройки лиги |
| `PATCH` | `/admin/settings` | ✅ | Обновить настройки. Body: `{ "values": { "k1": N, "k2": N, ... } }` |

## Примеры

### Вход
```bash
curl -X POST http://localhost:8081/admin/login \
  -H "Content-Type: application/json" \
  -d '{"login":"admin","password":"admin"}'
```

### Список игроков
```bash
curl http://localhost:8081/admin/players \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Бан игрока
```bash
curl -X POST http://localhost:8081/admin/players/42/ban \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Обновление настроек
```bash
curl -X PATCH http://localhost:8081/admin/settings \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"values":{"k1":2,"lp_win":5}}'
```
