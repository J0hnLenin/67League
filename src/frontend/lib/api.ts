// В dev-режиме Next.js проксирует /api/v1/* на бэкенд (см. next.config.ts).
// В production фронт и бэкенд на одном домене или проксируются через nginx.
const API_BASE = "/api/v1";

export interface ApiError {
  message: string;
  code?: string;
}

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token = typeof window !== "undefined" ? localStorage.getItem("token") : null;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...((options.headers as Record<string, string>) || {}),
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.message || `HTTP ${res.status}`);
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return res.json() as Promise<T>;
}

export const api = {
  // Auth
  register: (data: RegisterInput) =>
    request<{ token: string; player_id: number; nickname: string }>("/auth/register", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  login: (data: LoginInput) =>
    request<{ token: string; player: Player }>("/auth/login", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  // Player
  me: () => request<Player>("/me"),
  myBadges: () => request<PlayerBadge[]>("/me/badges"),
  myTitles: () => request<Title[]>("/me/titles"),
  myStoneHistory: () => request<StoneTransaction[]>("/me/stones/history"),
  myMatches: () => request<Match[]>("/me/matches"),
  searchPlayers: (query: string) =>
    request<Array<{ id: number; full_name: string; nickname: string }>>(`/players?q=${encodeURIComponent(query)}`),

  // Clubs
  clubs: () => request<Club[]>("/clubs"),
  club: (id: string) => request<Club>(`/clubs/${id}`),
  clubBadges: (id: string) => request<ClubBadge[]>(`/clubs/${id}/badges`),
  createClub: (data: CreateClubInput) =>
    request<Club>("/clubs", { method: "POST", body: JSON.stringify(data) }),
  joinClub: (id: string) =>
    request<void>(`/clubs/${id}/join`, { method: "POST" }),
  leaveClub: (id: string) =>
    request<void>(`/clubs/${id}/leave`, { method: "POST" }),

  // Seasons
  seasons: () => request<Season[]>("/seasons"),
  currentSeason: () => request<Season>("/seasons/current"),
  season: (id: string) => request<Season>(`/seasons/${id}`),
  joinSeason: (id: string) =>
    request<void>(`/seasons/${id}/join`, { method: "POST" }),

  // Divisions
  division: (id: string) => request<Division>(`/divisions/${id}`),
  divisionMembers: (id: string) => request<DivisionMember[]>(`/divisions/${id}/members`),

  // Matches
  match: (id: string) => request<Match>(`/matches/${id}`),
  createMatch: (data: CreateMatchInput) =>
    request<Match>("/matches", { method: "POST", body: JSON.stringify(data) }),
  acceptMatch: (id: string) =>
    request<void>(`/matches/${id}/accept`, { method: "POST" }),
  startMatch: (id: string) =>
    request<void>(`/matches/${id}/start`, { method: "POST" }),
  submitResult: (id: string, data: SubmitResultInput) =>
    request<void>(`/matches/${id}/result`, { method: "POST", body: JSON.stringify(data) }),
  confirmMatch: (id: string) =>
    request<void>(`/matches/${id}/confirm`, { method: "POST" }),
  cancelMatch: (id: string) =>
    request<void>(`/matches/${id}/cancel`, { method: "POST" }),
};

// ---- Types --------------------------------------------------------------------

export interface Player {
  id: string;
  full_name: string;
  email: string;
  nickname: string;
  city: string;
  rating: number;
  club_id?: string;
  stones: number;
  lp_total: number;
  status: "active" | "inactive" | "banned";
  active_title_id?: string;
  created_at: string;
}

export interface RegisterInput {
  full_name: string;
  email: string;
  password: string;
  nickname: string;
  city: string;
  rating?: number;
  kgs_nick?: string;
  ogs_nick?: string;
  fox_nick?: string;
  rfg_id?: string;
  sputnik_id?: string;
  club_id?: string;
}

export interface LoginInput {
  email: string;
  password: string;
}

export interface Club {
  id: string;
  name: string;
  city?: string;
  university?: string;
  logo_url?: string;
  created_by: string;
  status: "active" | "archived";
  created_at: string;
}

export interface CreateClubInput {
  name: string;
  city?: string;
  university?: string;
}

export interface Season {
  id: string;
  number: number;
  title: string;
  start_date: string;
  end_date: string;
  status: "announced" | "registration" | "active" | "closing" | "finished" | "archived";
  settings: SeasonSettings;
  created_at: string;
}

export interface SeasonSettings {
  k1: number;
  k2: number;
  k3: number;
  p: number;
  n: number;
  t: number;
  challenge_accept_days: number;
  match_cancel_days: number;
  season_duration_days: number;
  closing_days: number;
  lp_win: number;
  lp_loss: number;
  streak_bonus_3: number;
  streak_bonus_5: number;
  streak_bonus_67: number;
  stones_win: number;
  stones_loss: number;
  stones_club_top1: number;
  stones_super_title: number;
  stones_per_extra_match: number;
}

export interface Division {
  id: string;
  season_id: string;
  code: string;
  alias?: string;
  rating_min: number;
  rating_max: number;
  order: number;
  status: "active" | "finished";
}

export interface DivisionMember {
  id: string;
  division_id: string;
  player_id: string;
  lp: number;
  wins: number;
  losses: number;
  matches_played: number;
  streak_current: number;
  streak_max: number;
  final_place?: number;
  status: "active" | "excluded";
  player?: Player;
}

export interface Match {
  id: string;
  season_id: string;
  division_id: string;
  black_player_id: string;
  white_player_id: string;
  created_by: string;
  status: "proposed" | "accepted" | "playing" | "pending_confirmation" | "confirmed" | "cancelled";
  server: "ogs" | "kgs" | "fox" | "offline";
  time_control?: string;
  handicap: number;
  komi: number;
  winner_id?: string;
  score_diff?: number;
  sgf_url?: string;
  link?: string;
  result_submitted_by?: string;
  confirmed_by?: string;
  cancel_reason?: string;
  created_at: string;
  accepted_at?: string;
  started_at?: string;
  finished_at?: string;
  confirmed_at?: string;
  cancelled_at?: string;
}

export interface CreateMatchInput {
  season_id: string;
  division_id: string;
  black_player_id: string;
  white_player_id: string;
  server: "ogs" | "kgs" | "fox" | "offline";
  time_control?: string;
  handicap?: number;
  komi?: number;
}

export interface SubmitResultInput {
  winner_id: string;
  score_diff?: number;
  sgf_url?: string;
  link?: string;
}

export interface Badge {
  id: string;
  code: string;
  icon: string;
  name: string;
  description: string;
  type: "personal" | "seasonal" | "club" | "super";
  rarity: "common" | "rare" | "epic" | "legendary" | "6-7";
  title_id?: string;
}

export interface PlayerBadge {
  id: string;
  player_id: string;
  badge_id: string;
  season_id?: string;
  awarded_at: string;
  is_active: boolean;
  badge?: Badge;
}

export interface ClubBadge {
  id: string;
  club_id: string;
  badge_id: string;
  season_id?: string;
  awarded_at: string;
  badge?: Badge;
}

export interface Title {
  id: string;
  code: string;
  name: string;
  badge_id?: string;
}

export interface StoneTransaction {
  id: string;
  player_id: string;
  amount: number;
  reason: string;
  ref_id?: string;
  created_at: string;
}
