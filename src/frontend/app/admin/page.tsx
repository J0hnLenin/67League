"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { adminApi } from "@/lib/admin-api";

export default function AdminDashboardPage() {
  const router = useRouter();
  const [counts, setCounts] = useState({ players: 0, clubs: 0, matches: 0, seasons: 0 });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    const token = localStorage.getItem("admin_token");
    if (!token) {
      router.push("/admin/login");
      return;
    }

    Promise.all([
      adminApi.listPlayers().catch(() => []),
      adminApi.listClubs().catch(() => []),
      adminApi.listMatches().catch(() => []),
      adminApi.listSeasons().catch(() => []),
    ])
      .then(([players, clubs, matches, seasons]) => {
        setCounts({
          players: players.length,
          clubs: clubs.length,
          matches: matches.length,
          seasons: seasons.length,
        });
      })
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

        {/* Счётчики */}
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <CounterCard href="/admin/players" icon="👥" title="Игроки" count={counts.players} loading={loading} color="bg-blue-500" />
          <CounterCard href="/admin/clubs" icon="🏛️" title="Клубы" count={counts.clubs} loading={loading} color="bg-green-500" />
          <CounterCard href="/admin/matches" icon="⚔️" title="Партии" count={counts.matches} loading={loading} color="bg-orange-500" />
          <CounterCard href="/admin/seasons" icon="📅" title="Сезоны" count={counts.seasons} loading={loading} color="bg-purple-500" />
        </div>

        {/* Навигация */}
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
          <AdminCard href="/admin/seasons" icon="📅" title="Сезоны" desc="Создание, управление, ротация" />
          <AdminCard href="/admin/players" icon="👥" title="Игроки" desc="Список, бан, разбан, редактирование" />
          <AdminCard href="/admin/clubs" icon="🏛️" title="Клубы" desc="Список, модерация" />
          <AdminCard href="/admin/matches" icon="⚔️" title="Партии" desc="Отмена, корректировка" />
          <AdminCard href="/admin/settings" icon="⚙️" title="Настройки" desc="Константы лиги" />
        </div>
      </div>
    </main>
  );
}

function CounterCard({ href, icon, title, count, loading, color }: { href: string; icon: string; title: string; count: number; loading: boolean; color: string }) {
  return (
    <Link href={href}>
      <div className="bg-white rounded-2xl shadow-sm p-6 border border-gray-100 hover:shadow-md transition-shadow">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-gray-500 text-sm">{title}</p>
            <p className="text-3xl font-bold text-gray-900 mt-1">
              {loading ? "—" : count}
            </p>
          </div>
          <div className={`w-12 h-12 rounded-xl ${color} text-white flex items-center justify-center text-2xl`}>
            {icon}
          </div>
        </div>
      </div>
    </Link>
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
