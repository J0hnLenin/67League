-- ============================================================
-- 6-7 Лига: начальная схема
-- ============================================================

-- Игроки (без FK на клубы — добавим ниже через ALTER)
CREATE TABLE players (
    id              BIGSERIAL PRIMARY KEY,
    full_name       TEXT NOT NULL,
    email           TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    nickname        TEXT NOT NULL UNIQUE,
    kgs_nick        TEXT NOT NULL DEFAULT '',
    ogs_nick        TEXT NOT NULL DEFAULT '',
    fox_nick        TEXT NOT NULL DEFAULT '',
    rfg_id          TEXT NOT NULL DEFAULT '',
    sputnik_id      TEXT NOT NULL DEFAULT '',
    city            TEXT NOT NULL,
    rating          INTEGER NOT NULL DEFAULT 0,
    club_id         BIGINT,
    stones          NUMERIC(14, 2) NOT NULL DEFAULT 0,
    lp_total        NUMERIC(14, 2) NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'active',
    active_title_id BIGINT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Админы (отдельная сущность)
CREATE TABLE admins (
    id            BIGSERIAL PRIMARY KEY,
    login         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Клубы
CREATE TABLE clubs (
    id           BIGSERIAL PRIMARY KEY,
    name         TEXT NOT NULL UNIQUE,
    city         TEXT NOT NULL DEFAULT '',
    university   TEXT NOT NULL DEFAULT '',
    logo_url     TEXT NOT NULL DEFAULT '',
    created_by   BIGINT REFERENCES players(id) ON DELETE SET NULL,
    status       TEXT NOT NULL DEFAULT 'active',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- FK players.club_id → clubs.id добавляем после создания clubs
ALTER TABLE players
    ADD CONSTRAINT players_club_id_fkey
    FOREIGN KEY (club_id) REFERENCES clubs(id) ON DELETE SET NULL;

CREATE INDEX idx_players_club ON players(club_id);
CREATE INDEX idx_players_status ON players(status);

-- Сезоны
CREATE TABLE seasons (
    id         BIGSERIAL PRIMARY KEY,
    number     INTEGER NOT NULL UNIQUE,
    title      TEXT NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date   TIMESTAMPTZ NOT NULL,
    status     TEXT NOT NULL DEFAULT 'announced',
    settings   JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Дивизионы сезона
CREATE TABLE divisions (
    id           BIGSERIAL PRIMARY KEY,
    season_id    BIGINT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
    code         TEXT NOT NULL,
    alias        TEXT NOT NULL DEFAULT '',
    rating_min   INTEGER NOT NULL DEFAULT 0,
    rating_max   INTEGER NOT NULL DEFAULT 0,
    "order"      INTEGER NOT NULL,
    status       TEXT NOT NULL DEFAULT 'active',
    rotation_up  INTEGER NOT NULL DEFAULT 0,
    rotation_down INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(season_id, code)
);
CREATE INDEX idx_divisions_season ON divisions(season_id);

-- Участие в дивизионе
CREATE TABLE division_members (
    id               BIGSERIAL PRIMARY KEY,
    division_id      BIGINT NOT NULL REFERENCES divisions(id) ON DELETE CASCADE,
    player_id        BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    lp               NUMERIC(14, 2) NOT NULL DEFAULT 0,
    wins             INTEGER NOT NULL DEFAULT 0,
    losses           INTEGER NOT NULL DEFAULT 0,
    matches_played   INTEGER NOT NULL DEFAULT 0,
    streak_current   INTEGER NOT NULL DEFAULT 0,
    streak_max       INTEGER NOT NULL DEFAULT 0,
    streaks_broken   INTEGER NOT NULL DEFAULT 0,
    final_place      INTEGER,
    status           TEXT NOT NULL DEFAULT 'active',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(division_id, player_id)
);
CREATE INDEX idx_members_player ON division_members(player_id);
CREATE INDEX idx_members_division ON division_members(division_id);

-- Партии
CREATE TABLE matches (
    id                 BIGSERIAL PRIMARY KEY,
    season_id          BIGINT NOT NULL REFERENCES seasons(id) ON DELETE RESTRICT,
    division_id        BIGINT NOT NULL REFERENCES divisions(id) ON DELETE RESTRICT,
    black_player_id    BIGINT NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    white_player_id    BIGINT NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    created_by         BIGINT NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    status             TEXT NOT NULL DEFAULT 'proposed',
    server             TEXT NOT NULL,
    time_control       TEXT NOT NULL DEFAULT '',
    handicap           INTEGER NOT NULL DEFAULT 0,
    komi               NUMERIC(6, 2) NOT NULL DEFAULT 0,
    winner_id          BIGINT REFERENCES players(id) ON DELETE SET NULL,
    score_diff         NUMERIC(6, 2),
    sgf_url            TEXT NOT NULL DEFAULT '',
    link               TEXT NOT NULL DEFAULT '',
    result_submitted_by BIGINT REFERENCES players(id) ON DELETE SET NULL,
    confirmed_by        BIGINT REFERENCES players(id) ON DELETE SET NULL,
    cancel_reason       TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    accepted_at         TIMESTAMPTZ,
    started_at          TIMESTAMPTZ,
    finished_at         TIMESTAMPTZ,
    confirmed_at        TIMESTAMPTZ,
    cancelled_at        TIMESTAMPTZ
);
CREATE INDEX idx_matches_season ON matches(season_id);
CREATE INDEX idx_matches_division ON matches(division_id);
CREATE INDEX idx_matches_black ON matches(black_player_id);
CREATE INDEX idx_matches_white ON matches(white_player_id);
CREATE INDEX idx_matches_status ON matches(status);

-- Журнал событий партии
CREATE TABLE match_events (
    id         BIGSERIAL PRIMARY KEY,
    match_id   BIGINT NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    actor_id   BIGINT REFERENCES players(id) ON DELETE SET NULL,
    payload    JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_match_events_match ON match_events(match_id);

-- Бейджи (справочник)
CREATE TABLE badges (
    id          BIGSERIAL PRIMARY KEY,
    code        TEXT NOT NULL UNIQUE,
    icon        TEXT NOT NULL DEFAULT '',
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    type        TEXT NOT NULL,
    condition   JSONB NOT NULL DEFAULT '{}'::jsonb,
    rarity      TEXT NOT NULL,
    title_id    BIGINT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Титулы (справочник)
CREATE TABLE titles (
    id       BIGSERIAL PRIMARY KEY,
    code     TEXT NOT NULL UNIQUE,
    name     TEXT NOT NULL,
    badge_id BIGINT REFERENCES badges(id) ON DELETE SET NULL
);

-- Выданные бейджи игрокам
CREATE TABLE player_badges (
    id         BIGSERIAL PRIMARY KEY,
    player_id  BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    badge_id   BIGINT NOT NULL REFERENCES badges(id) ON DELETE RESTRICT,
    season_id  BIGINT REFERENCES seasons(id) ON DELETE SET NULL,
    awarded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_active  BOOLEAN NOT NULL DEFAULT true
);
CREATE INDEX idx_player_badges_player ON player_badges(player_id);
CREATE INDEX idx_player_badges_season ON player_badges(season_id);

-- Выданные бейджи клубам
CREATE TABLE club_badges (
    id         BIGSERIAL PRIMARY KEY,
    club_id    BIGINT NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    badge_id   BIGINT NOT NULL REFERENCES badges(id) ON DELETE RESTRICT,
    season_id  BIGINT REFERENCES seasons(id) ON DELETE SET NULL,
    awarded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(club_id, badge_id)
);

-- Супертитул «Чемпион 67»
CREATE TABLE super_titles (
    id         BIGSERIAL PRIMARY KEY,
    season_id  BIGINT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
    player_id  BIGINT NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    club_id    BIGINT NOT NULL REFERENCES clubs(id) ON DELETE RESTRICT,
    awarded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(season_id)
);

-- Транзакции камней
CREATE TABLE stone_transactions (
    id         BIGSERIAL PRIMARY KEY,
    player_id  BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    amount     NUMERIC(14, 2) NOT NULL,
    reason     TEXT NOT NULL,
    ref_id     TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_stone_tx_player ON stone_transactions(player_id);

-- Покупки
CREATE TABLE purchases (
    id         BIGSERIAL PRIMARY KEY,
    player_id  BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    item_code  TEXT NOT NULL,
    price      NUMERIC(14, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Уведомления
CREATE TABLE notifications (
    id         BIGSERIAL PRIMARY KEY,
    player_id  BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    type       TEXT NOT NULL,
    payload    JSONB NOT NULL DEFAULT '{}'::jsonb,
    channel    TEXT NOT NULL,
    sent_at    TIMESTAMPTZ,
    read_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_notifications_player ON notifications(player_id);
CREATE INDEX idx_notifications_unsent ON notifications(id) WHERE sent_at IS NULL;

-- Глобальные настройки
CREATE TABLE site_settings (
    key        TEXT PRIMARY KEY,
    value      JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);