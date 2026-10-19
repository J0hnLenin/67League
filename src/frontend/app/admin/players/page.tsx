"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { adminApi } from "@/lib/admin-api";
import { Player } from "@/lib/api";

export default function AdminPlayersPage() {
  const router = useRouter();
  const [players, setPlayers] = useState<Player[]>([]);
  const [filtered, setFiltered] = useState<Player[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");

  useEffect(() => {
    const token = localStorage.getItem("admin_token");
    if (!token) {
      router.push("/admin/login");
      return;
    }
    loadPlayers();
  }, [router]);

  useEffect(() => {
    let result = players;
    if (search) {
      const q = search.toLowerCase();
      result = result.filter(
        (p) =>
          p.nickname.toLowerCase().includes(q) ||
          p.full_name.toLowerCase().includes(q) ||
          p.email?.toLowerCase().includes(q)
      );
    }
    if (statusFilter !== "all") {
      result = result.filter((p) => p.status === statusFilter);
    }
    setFiltered(result);
  }, [players, search, statusFilter]);

  function loadPlayers() {
    setLoading(true);
    adminApi
      .listPlayers()
      .then((data) => {
        setPlayers(data);
        setFiltered(data);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }

  async function toggleBan(p: Player) {
    try {
      if (p.status === "banned") {
        await adminApi.unbanPlayer(String(p.id));
      } else {
        await adminApi.banPlayer(String(p.id));
      }
      loadPlayers();
    } catch (err: any) {
      setError(err.message);
    }
  }

  return (
    <main className="min-h-screen bg-gray-50">
      <header className="bg-gray-900 text-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <Link href="/" className="text-xl font-bold">6-7 Лига</Link>
          <div className="flex items-center gap-4">
            <Link href="/admin" className="text-gray-400 hover:text-white text-sm">← Назад</Link>
            <button
              onClick={() => { localStorage.removeItem("admin_token"); window.location.href = "/"; }}
              className="text-gray-400 hover:text-white text-sm"
            >
              Выйти
            </button>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Игроки</h1>
          <span className="text-gray-500">{filtered.length} из {players.length}</span>
        </div>

        {error && <p className="text-red-600 mb-4">{error}</p>}

        <div className="flex gap-4 mb-6">
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Поиск по нику, имени, email..."
            className="flex-1 px-4 py-2 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none"
          />
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-4 py-2 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none"
          >
            <option value="all">Все статусы</option>
            <option value="active">Активные</option>
            <option value="inactive">Неактивные</option>
            <option value="banned">Забаненные</option>
          </select>
        </div>

        {loading ? (
          <p className="text-gray-500">Загрузка...</p>
        ) : (
          <div className="bg-white rounded-2xl shadow-sm overflow-hidden">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">ID</th>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Ник</th>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Имя</th>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Email</th>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Город</th>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Рейтинг</th>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Статус</th>
                  <th className="text-right px-4 py-3 text-sm font-medium text-gray-500">Действия</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {filtered.map((p) => (
                  <tr key={p.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm text-gray-600">{p.id}</td>
                    <td className="px-4 py-3 text-sm font-medium text-gray-900">
                      <Link href={`/admin/players/${p.id}`} className="text-primary-600 hover:underline">
                        {p.nickname}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600">{p.full_name}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">{p.email}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">{p.city}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">{p.rating}</td>
                    <td className="px-4 py-3">
                      <StatusBadge status={p.status} />
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        onClick={() => toggleBan(p)}
                        className={`text-sm ${p.status === "banned" ? "text-green-600 hover:text-green-800" : "text-red-600 hover:text-red-800"}`}
                      >
                        {p.status === "banned" ? "Разбанить" : "Забанить"}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {filtered.length === 0 && (
              <p className="text-gray-500 text-center py-12">Игроки не найдены.</p>
            )}
          </div>
        )}
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
