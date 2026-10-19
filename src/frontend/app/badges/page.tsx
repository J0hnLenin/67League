"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, PlayerBadge } from "@/lib/api";

export default function BadgesPage() {
  const [badges, setBadges] = useState<PlayerBadge[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    api.myBadges()
      .then(setBadges)
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
        <h1 className="text-3xl font-bold text-gray-900 mb-8">Мои бейджи</h1>

        {loading && <p className="text-gray-500">Загрузка...</p>}
        {error && <p className="text-red-600">{error}</p>}

        {!loading && !error && (
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {badges.map((pb) => (
              <div key={pb.id} className="bg-white rounded-2xl shadow-sm p-6 border border-gray-100">
                <div className="text-4xl mb-3">{pb.badge?.icon || "🏅"}</div>
                <h3 className="font-bold text-gray-900">{pb.badge?.name || "Бейдж"}</h3>
                <p className="text-gray-500 text-sm mt-1">{pb.badge?.description || ""}</p>
                <p className="text-gray-400 text-xs mt-3">
                  Получен: {new Date(pb.awarded_at).toLocaleDateString("ru-RU")}
                </p>
              </div>
            ))}
            {badges.length === 0 && (
              <p className="text-gray-500 col-span-full text-center py-12">
                Пока нет бейджей. Играйте и зарабатывайте достижения!
              </p>
            )}
          </div>
        )}
      </div>
    </main>
  );
}
