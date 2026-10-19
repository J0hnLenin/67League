"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { adminApi } from "@/lib/admin-api";

export default function AdminDashboardPage() {
  const router = useRouter();
  const [players, setPlayers] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    const token = localStorage.getItem("admin_token");
    if (!token) {
      router.push("/admin/login");
      return;
    }

    adminApi.listPlayers()
      .then((res) => setPlayers(res.players || []))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [router]);

  return (
    <main className="min-h-screen bg-gray-50">
      <header className="bg-gray-900 text-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <Link href="/" className="text-xl font-bold">6-7 Лига</Link>
          <span className="text-gray-400 text-sm">Админ-панель</span>
          <button
            onClick={() => {
              localStorage.removeItem("admin_token");
              window.location.href = "/";
            }}
            className="text-gray-400 hover:text-white text-sm"
          >
            Выйти
          </button>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">Администрирование</h1>

        {error && <p className="text-red-600 mb-4">{error}</p>}

        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <AdminCard href="/admin/seasons" icon="📅" title="Сезоны" desc="Создание, управление, ротация" />
          <AdminCard href="/admin/players" icon="👥" title="Игроки" desc="Список, бан, разбан" />
          <AdminCard href="/admin/matches" icon="⚔️" title="Партии" desc="Отмена, корректировка" />
          <AdminCard href="/admin/settings" icon="⚙️" title="Настройки" desc="Константы лиги" />
        </div>

        <div className="bg-white rounded-2xl shadow-sm p-6">
          <h2 className="text-xl font-bold text-gray-900 mb-4">Игроки ({players.length})</h2>
          {loading ? (
            <p className="text-gray-500">Загрузка...</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">ID</th>
                    <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Ник</th>
                    <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Email</th>
                    <th className="text-left px-4 py-3 text-sm font-medium text-gray-500">Статус</th>
                    <th className="text-right px-4 py-3 text-sm font-medium text-gray-500">Действия</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {players.slice(0, 10).map((p) => (
                    <tr key={p.id} className="hover:bg-gray-50">
                      <td className="px-4 py-3 text-sm text-gray-600">{p.id}</td>
                      <td className="px-4 py-3 text-sm font-medium text-gray-900">{p.nickname}</td>
                      <td className="px-4 py-3 text-sm text-gray-600">{p.email}</td>
                      <td className="px-4 py-3">
                        <span className={`text-xs px-2 py-1 rounded-full ${
                          p.status === "active" ? "bg-green-100 text-green-700" :
                          p.status === "banned" ? "bg-red-100 text-red-700" :
                          "bg-gray-100 text-gray-600"
                        }`}>
                          {p.status}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-right">
                        {p.status !== "banned" ? (
                          <button
                            onClick={() => adminApi.banPlayer(String(p.id)).then(() => window.location.reload())}
                            className="text-red-600 hover:text-red-800 text-sm"
                          >
                            Забанить
                          </button>
                        ) : (
                          <button
                            onClick={() => adminApi.unbanPlayer(String(p.id)).then(() => window.location.reload())}
                            className="text-green-600 hover:text-green-800 text-sm"
                          >
                            Разбанить
                          </button>
                        )}
                      </td>
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

function AdminCard({ href, icon, title, desc }: { href: string; icon: string; title: string; desc: string }) {
  return (
    <Link href={href}>
      <div className="bg-white rounded-2xl shadow-sm p-6 hover:shadow-md transition-shadow border border-gray-100">
        <div className="text-3xl mb-3">{icon}</div>
        <h3 className="font-bold text-gray-900">{title}</h3>
        <p className="text-gray-500 text-sm mt-1">{desc}</p>
      </div>
    </Link>
  );
}
