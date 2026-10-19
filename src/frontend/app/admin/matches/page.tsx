"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { adminApi } from "@/lib/admin-api";
import { Match } from "@/lib/api";

export default function AdminMatchesPage() {
  const router = useRouter();
  const [matches, setMatches] = useState<Match[]>([]);
  const [filtered, setFiltered] = useState<Match[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");

  useEffect(() => {
    const token = localStorage.getItem("admin_token");
    if (!token) {
      router.push("/admin/login");
      return;
    }
    loadMatches();
  }, [router]);

  useEffect(() => {
    if (statusFilter === "all") {
      setFiltered(matches);
    } else {
      setFiltered(matches.filter((m) => m.status === statusFilter));
    }
  }, [matches, statusFilter]);

  function loadMatches() {
    setLoading(true);
    adminApi
      .listMatches()
      .then((data) => {
        setMatches(data);
        setFiltered(data);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }

  async function cancelMatch(id: string) {
    if (!confirm("Отменить партию?")) return;
    try {
      await adminApi.cancelMatch(id);
      loadMatches();
    } catch (err: any) {
      setError(err.message);
    }
  }

  const statusLabels: Record<string, string> = {
    proposed: "Выбор",
    accepted: "Принят",
    playing: "Идёт",
    pending_confirmation: "Ожидает",
    confirmed: "Подтверждён",
    cancelled: "Отменён",
  };

  const statusColors: Record<string, string> = {
    proposed: "bg-blue-100 text-blue-700",
    accepted: "bg-yellow-100 text-yellow-700",
    playing: "bg-primary-100 text-primary-700",
    pending_confirmation: "bg-orange-100 text-orange-700",
    confirmed: "bg-green-100 text-green-700",
    cancelled: "bg-gray-100 text-gray-500",
  };

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
          <h1 className="text-3xl font-bold text-gray-900">Партии</h1>
          <div className="flex items-center gap-4">
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="px-4 py-2 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none"
            >
              <option value="all">Все статусы</option>
              <option value="proposed">Вызов</option>
              <option value="accepted">Принят</option>
              <option value="playing">Идёт</option>
              <option value="pending_confirmation">Ожидает</option>
              <option value="confirmed">Подтверждён</option>
              <option value="cancelled">Отменён</option>
            </select>
            <span className="text-gray-500">{filtered.length} из {matches.length}</span>
          </div>
        </div>

        {error && <p className="text-red-600 mb-4">{error}</p>}

        {loading ? (
          <p className="text-gray-500">Загрузка...</p>
        ) : (
          <div className="bg-white rounded-2xl shadow-sm overflow-hidden">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">ID</th>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Статус</th>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Сервер</th>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Чёрные</th>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Белые</th>
                  <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Победитель</th>
                  <th className="text-right px-4 py-3 text-sm font-medium text-gray-500">Действия</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {filtered.map((m) => (
                  <tr key={m.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-sm text-gray-600">{m.id}</td>
                    <td className="px-4 py-3">
                      <span className={`text-xs px-2 py-1 rounded-full ${statusColors[m.status]}`}>
                        {statusLabels[m.status] || m.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600 uppercase">{m.server}</td>
                    <td className="px-4 py-3 text-sm text-gray-900">{m.black_player_id}</td>
                    <td className="px-4 py-3 text-sm text-gray-900">{m.white_player_id}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">{m.winner_id || "—"}</td>
                    <td className="px-4 py-3 text-right">
                      {m.status !== "cancelled" && m.status !== "confirmed" && (
                        <button
                          onClick={() => cancelMatch(m.id)}
                          className="text-red-600 hover:text-red-800 text-sm"
                        >
                          Отменить
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {filtered.length === 0 && (
              <p className="text-gray-500 text-center py-12">Партии не найдены.</p>
            )}
          </div>
        )}
      </div>
    </main>
  );
}
