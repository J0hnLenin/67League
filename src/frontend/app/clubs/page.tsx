"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, Club } from "@/lib/api";

export default function ClubsPage() {
  const [clubs, setClubs] = useState<Club[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    api.clubs()
      .then(setClubs)
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
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Клубы</h1>
          <Link
            href="/clubs/create"
            className="bg-primary-600 text-white px-6 py-2 rounded-lg font-medium hover:bg-primary-700 transition-colors"
          >
            Создать клуб
          </Link>
        </div>

        {loading && <p className="text-gray-500">Загрузка...</p>}
        {error && <p className="text-red-600">{error}</p>}

        {!loading && !error && (
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {clubs.map((club) => (
              <ClubCard key={club.id} club={club} />
            ))}
            {clubs.length === 0 && (
              <p className="text-gray-500 col-span-full text-center py-12">
                Пока нет клубов. Станьте первым!
              </p>
            )}
          </div>
        )}
      </div>
    </main>
  );
}

function ClubCard({ club }: { club: Club }) {
  return (
    <Link href={`/clubs/${club.id}`}>
      <div className="bg-white rounded-2xl shadow-sm p-6 hover:shadow-md transition-shadow border border-gray-100">
        <div className="flex items-start gap-4">
          <div className="w-12 h-12 bg-primary-100 rounded-xl flex items-center justify-center text-2xl flex-shrink-0">
            🏛️
          </div>
          <div className="min-w-0">
            <h3 className="font-bold text-gray-900 truncate">{club.name}</h3>
            {club.city && <p className="text-gray-500 text-sm">{club.city}</p>}
            {club.university && (
              <p className="text-gray-400 text-sm truncate">{club.university}</p>
            )}
          </div>
        </div>
        <div className="mt-4 flex items-center gap-2">
          <span
            className={`text-xs px-2 py-1 rounded-full ${
              club.status === "active"
                ? "bg-green-100 text-green-700"
                : "bg-gray-100 text-gray-600"
            }`}
          >
            {club.status === "active" ? "Активен" : "В архиве"}
          </span>
        </div>
      </div>
    </Link>
  );
}
