"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, Match } from "@/lib/api";

export default function MatchesPage() {
  const [matches, setMatches] = useState<Match[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    api.myMatches()
      .then(setMatches)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  return (
    <main className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <Link href="/" className="text-xl font-bold text-primary-700">6-7 Лига</Link>
          <nav className="flex items-center gap-4">
            <Link href="/dashboard" className="text-gray-600 hover:text-primary-600">Личный кабинет</Link>
          </nav>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Мои партии</h1>
          <Link href="/matches/new" className="bg-primary-600 text-white px-6 py-2 rounded-lg font-medium hover:bg-primary-700 transition-colors">
            Создать вызов
          </Link>
        </div>

        {loading && <p className="text-gray-500">Загрузка...</p>}
        {error && <p className="text-red-600">{error}</p>}

        {!loading && !error && (
          <div className="space-y-4">
            {matches.map((m) => (
              <MatchRow key={m.id} match={m} />
            ))}
            {matches.length === 0 && (
              <p className="text-gray-500 text-center py-12">Пока нет партий. Создайте первый вызов!</p>
            )}
          </div>
        )}
      </div>
    </main>
  );
}

function MatchRow({ match }: { match: Match }) {
  const statusLabels: Record<string, string> = {
    proposed: "🎯 Вызов",
    accepted: "✅ Принят",
    playing: "⚔️ Идёт",
    pending_confirmation: "⏳ Ожидает подтверждения",
    confirmed: "✓ Завершён",
    cancelled: "✗ Отменён",
  };

  return (
    <Link href={`/matches/${match.id}`}>
      <div className="bg-white rounded-xl shadow-sm p-6 hover:shadow-md transition-shadow border border-gray-100">
        <div className="flex items-center justify-between">
          <div>
            <p className="font-semibold text-gray-900">
              Чёрные #{match.black_player_id} vs Белые #{match.white_player_id}
            </p>
            <p className="text-gray-500 text-sm mt-1">
              {new Date(match.created_at).toLocaleDateString("ru-RU")} · {match.server.toUpperCase()}
            </p>
          </div>
          <span className="text-sm px-3 py-1 rounded-full bg-gray-100 text-gray-700 font-medium">
            {statusLabels[match.status] || match.status}
          </span>
        </div>
      </div>
    </Link>
  );
}
