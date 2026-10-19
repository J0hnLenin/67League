// В dev-режиме Next.js проксирует /admin/* на админ-бэкенд (см. next.config.ts).
const ADMIN_API_BASE = "/admin";

async function adminRequest<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token = typeof window !== "undefined" ? localStorage.getItem("admin_token") : null;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...((options.headers as Record<string, string>) || {}),
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${ADMIN_API_BASE}${path}`, {
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

export const adminApi = {
  login: (data: { email: string; password: string }) =>
    adminRequest<{ token: string; player_id: number; nickname: string }>("/login", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  // Seasons
  createSeason: (data: { number: number; title: string; start_date: string; end_date: string }) =>
    adminRequest<{ season_id: number }>("/seasons", { method: "POST", body: JSON.stringify(data) }),

  startRegistration: (id: string) =>
    adminRequest<void>(`/seasons/${id}/start-registration`, { method: "POST" }),
  startSeason: (id: string) =>
    adminRequest<void>(`/seasons/${id}/start`, { method: "POST" }),
  closeSeason: (id: string) =>
    adminRequest<void>(`/seasons/${id}/close`, { method: "POST" }),
  finishSeason: (id: string) =>
    adminRequest<void>(`/seasons/${id}/finish`, { method: "POST" }),
  archiveSeason: (id: string) =>
    adminRequest<void>(`/seasons/${id}/archive`, { method: "POST" }),

  calculateResults: (id: string) =>
    adminRequest<void>(`/seasons/${id}/calculate-results`, { method: "POST" }),
  applyRotation: (id: string) =>
    adminRequest<void>(`/seasons/${id}/apply-rotation`, { method: "POST" }),
  awardSeasonBadges: (id: string) =>
    adminRequest<void>(`/seasons/${id}/award-season-badges`, { method: "POST" }),
  awardSuperTitle: (id: string) =>
    adminRequest<void>(`/seasons/${id}/award-super-title`, { method: "POST" }),

  // Players
  listPlayers: () =>
    adminRequest<{ players: Array<{ id: number; full_name: string; nickname: string; email: string; status: string }> }>("/players"),
  banPlayer: (id: string) =>
    adminRequest<void>(`/players/${id}/ban`, { method: "POST" }),
  unbanPlayer: (id: string) =>
    adminRequest<void>(`/players/${id}/unban`, { method: "POST" }),

  // Matches
  cancelMatch: (id: string) =>
    adminRequest<void>(`/matches/${id}/cancel`, { method: "POST" }),

  // Settings
  getSettings: () =>
    adminRequest<Record<string, any>>("/settings"),
  patchSettings: (data: Record<string, any>) =>
    adminRequest<void>("/settings", { method: "PATCH", body: JSON.stringify(data) }),
};
