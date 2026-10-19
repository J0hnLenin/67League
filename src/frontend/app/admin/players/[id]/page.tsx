"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useParams } from "next/navigation";
import { adminApi } from "@/lib/admin-api";
import { Player, Match } from "@/lib/api";

export default function AdminPlayerDetailPage() {
  const router = useRouter();
  const params = useParams();
  const playerId = params.id as string;

  const [player, setPlayer] = useState<Player | null>(null);
  const [matches, setMatches] = useState<Match[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);

  // Edit form state
  const [fullName, setFullName] = useState("");
  const [city, setCity] = useState("");
  const [rating, setRating] = useState(0);
  const [status, setStatus] = useState<"active" | "inactive" | "banned">("active");

  useEffect(() => {
    const token = localStorage.getItem("admin_token");
    if (!token) {
      router.push("/admin/login");
      return;
    }
    loadData();
  }, [router, playerId]);

  function loadData() {
    setLoading(true);
    Promise.all([
      adminApi.getPlayer(playerId),
      adminApi.playerMatches(playerId),
    ])
      .then(([p, m]) => {
        setPlayer(p);
        setMatches(m);
        setFullName(p.full_name);
        setCity(p.city);
        setRating(p.rating);
        setStatus(p.status);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }

  async function handleSave() {
    setSaving(true);
    try {
      const updated = await adminApi.updatePlayer(playerId, {
        full_name: fullName,
        city,
        rating,
        status,
      });
      setPlayer(updated);
      setEditing(false);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  }

  async function toggleBan() {
    if (!player) return;
    try {
      if (player.status === "banned") {
        await adminApi.unbanPlayer(playerId);
      } else {
        await adminApi.banPlayer(playerId);
      }
      loadData();
    } catch (err: any) {
      setError(err.message);
    }
  }

  if (loading) {
    return (
      <main className="min-h-screen bg-gray-50 flex items-center justify-center">
        <p className="text-gray-500">Загрузка...</p>
      </main>
    );
  }

  if (!player) {
    return (
      <main className="min-h-screen bg-gray-50 flex items-center justify-center">
        <p className="text-gray-500">Игрок не найден</p>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-gray-50">
      <header className="bg-gray-900 text-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <Link href="/" className="text-xl font-bold">6-7 Лига</Link>
          <div className="flex items-center gap-4">
            <Link href="/admin/players" className="text-gray-400 hover:text-white text-sm">← К списку</Link>
            <button
              onClick={() => { localStorage.removeItem("admin_token"); window.location.href = "/"; }}
              className="text-gray-400 hover:text-white text-sm"
            >
              Выйти
            </button>
          </div>
        </div>
      </header>

      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {error && <p className="text-red-600 mb-4">{error}</p>}

        <div className="bg-white rounded-2xl shadow-sm p-6 mb-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">{player.nickname}</h1>
              <p className="text-gray-500">ID: {player.id}</p>
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => setEditing(!editing)}
                className="bg-gray-100 hover:bg-gray-200 text-gray-700 px-4 py-2 rounded-lg text-sm font-medium transition-colors"
              >
                {editing ? "Отмена" : "Редактировать"}
              </button>
              <button
                onClick={toggleBan}
                className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                  player.status === "banned"
                    ? "bg-green-100 hover:bg-green-200 text-green-700"
                    : "bg-red-100 hover:bg-red-200 text-red-700"
                }`}
              >
                {player.status === "banned" ? "Разбанить" : "Забанить"}
              </button>
            </div>
          </div>

          {editing ? (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Имя</label>
                  <input
                    type="text"
                    value={fullName}
                    onChange={(e) => setFullName(e.target.value)}
                    className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Город</label>
                  <input
                    type="text"
                    value={city}
                    onChange={(e) => setCity(e.target.value)}
                    className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Рейтинг</label>
                  <input
                    type="number"
                    value={rating}
                    onChange={(e) => setRating(Number(e.target.value))}
                    className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Статус</label>
                  <select
                    value={status}
                    onChange={(e) => setStatus(e.target.value as "active" | "inactive" | "banned")}
                    className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none"
                  >
                    <option value="active">Активен</option>
                    <option value="inactive">Неактивен</option>
                    <option value="banned">Забанен</option>
                  </select>
                </div>
              </div>
              <button
                onClick={handleSave}
                disabled={saving}
                className="bg-primary-600 text-white px-6 py-2 rounded-lg font-medium hover:bg-primary-700 transition-colors disabled:opacity-50"
              >
                {saving ? "Сохранение..." : "Сохранить"}
              </button>
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <p className="text-gray-500">Имя</p>
                <p className="font-medium text-gray-900">{player.full_name}</p>
              </div>
              <div>
                <p className="text-gray-500">Email</p>
                <p className="font-medium text-gray-900">{player.email}</p>
              </div>
              <div>
                <p className="text-gray-500">Город</p>
                <p className="font-medium text-gray-900">{player.city}</p>
              </div>
              <div>
                <p className="text-gray-500">Рейтинг</p>
                <p className="font-medium text-gray-900">{player.rating}</p>
              </div>
              <div>
                <p className="text-gray-500">Камни</p>
                <p className="font-medium text-gray-900">{player.stones}</p>
              </div>
              <div>
                <p className="text-gray-500">LP</p>
                <p className="font-medium text-gray-900">{player.lp_total}</p>
              </div>
              <div>
                <p className="text-gray-500">Статус</p>
                <StatusBadge status={player.status} />
              </div>
              <div>
                <p className="text-gray-500">Клуб</p>
                <p className="font-medium text-gray-900">{player.club_id || "—"}</p>
              </div>
            </div>
          )}
        </div>

        {/* Партии игрока */}
        <div className="bg-white rounded-2xl shadow-sm p-6">
          <h2 className="text-lg font-bold text-gray-900 mb-4">Партии ({matches.length})</h2>
          {matches.length === 0 ? (
            <p className="text-gray-500">Пока нет партий.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="text-left px-4 py-2 text-sm font-medium text-gray-500">ID</th>
                    <th className="text-left px-4 py-2 text-sm font-medium text-gray-500">Статус</th>
                    <th className="text-left px-4 py-2 text-sm font-medium text-gray-500">Сервер</th>
                    <th className="text-left px-4 py-2 text-sm font-medium text-gray-500">Победитель</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {matches.map((m) => (
                    <tr key={m.id} className="hover:bg-gray-50">
                      <td className="px-4 py-2 text-sm text-gray-600">{m.id}</td>
                      <td className="px-4 py-2 text-sm">{m.status}</td>
                      <td className="px-4 py-2 text-sm text-gray-600 uppercase">{m.server}</td>
                      <td className="px-4 py-2 text-sm text-gray-600">{m.winner_id || "—"}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </main>
  );
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    active: "bg-green-100 text-green-700",
    inactive: "bg-gray-100 text-gray-600",
    banned: "bg-red-100 text-red-700",
  };
  const labels: Record<string, string> = {
    active: "Активен",
    inactive: "Неактивен",
    banned: "Забанен",
  };
  return (
    <span className={`text-xs px-2 py-1 rounded-full ${colors[status] || "bg-gray-100 text-gray-600"}`}>
      {labels[status] || status}
    </span>
  );
}
