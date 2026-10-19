"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, Season } from "@/lib/api";

export default function SeasonsPage() {
  const [seasons, setSeasons] = useState<Season[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    api.seasons()
      .then(setSeasons)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  return (
    <main className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <Link href="/" className="text-xl font-bold text-primary-700">
            6-7 Лига
          </Link>
          <nav className="flex items-center gap-4">
            <Link href="/dashboard" className="text-gray-600 hover:text-primary-600">
              Личный кабинет
            </Link>
          </nav>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">Сезоны</h1>

        {loading && <p className="text-gray-500">Загрузка...</p>}
        {error && <p className="text-red-600">{error}</p>}

        {!loading && !error && (
          <div className="space-y-4">
            {seasons.map((season) => (
              <SeasonRow key={season.id} season={season} />
            ))}
            {seasons.length === 0 && (
              <p className="text-gray-500 text-center py-12">Пока нет сезонов.</p>
            )}
          </div>
        )}
      </div>
    </main>
  );
}

function SeasonRow({ season }: { season: Season }) {
  const statusColors: Record<string, string> = {
    announced: "bg-blue-100 text-blue-700",
    registration: "bg-green-100 text-green-700",
    active: "bg-primary-100 text-primary-700",
    closing: "bg-yellow-100 text-yellow-700",
    finished: "bg-gray-100 text-gray-600",
    archived: "bg-gray-100 text-gray-500",
  };

  const statusLabels: Record<string, string> = {
    announced: "Анонсирован",
    registration: "Регистрация",
    active: "Активен",
    closing: "Закрытие",
    finished: "Завершён",
    archived: "В архиве",
  };

  return (
    <Link href={`/seasons/${season.id}`}>
      <div className="bg-white rounded-xl shadow-sm p-6 hover:shadow-md transition-shadow border border-gray-100 flex items-center justify-between">
        <div>
          <h3 className="font-bold text-gray-900 text-lg">{season.title}</h3>
          <p className="text-gray-500 text-sm mt-1">
            {new Date(season.start_date).toLocaleDateString("ru-RU")} —{" "}
            {new Date(season.end_date).toLocaleDateString("ru-RU")}
          </p>
        </div>
        <span className={`text-sm px-3 py-1 rounded-full font-medium ${statusColors[season.status] || "bg-gray-100"}`}>
          {statusLabels[season.status] || season.status}
        </span>
      </div>
    </Link>
  );
}
