"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { adminApi } from "@/lib/admin-api";
import { Club } from "@/lib/api";

export default function AdminClubsPage() {
  const router = useRouter();
  const [clubs, setClubs] = useState<Club[]>([]);
  const [filtered, setFiltered] = useState<Club[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [search, setSearch] = useState("");

  useEffect(() => {
    const token = localStorage.getItem("admin_token");
    if (!token) {
      router.push("/admin/login");
      return;
    }
    loadClubs();
  }, [router]);

  useEffect(() => {
    if (search) {
      const q = search.toLowerCase();
      setFiltered(
        clubs.filter(
          (c) =>
            c.name.toLowerCase().includes(q) ||
            (c.city && c.city.toLowerCase().includes(q))
        )
      );
    } else {
      setFiltered(clubs);
    }
  }, [clubs, search]);

  function loadClubs() {
    setLoading(true);
    adminApi
      .listClubs()
      .then((data) => {
        setClubs(data);
        setFiltered(data);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
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
          <h1 className="text-3xl font-bold text-gray-900">Клубы</h1>
          <span className="text-gray-500">{filtered.length} из {clubs.length}</span>
        </div>

        {error && <p className="text-red-600 mb-4">{error}</p>}

        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Поиск по названию, городу..."
          className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none mb-6"
        />

        {loading ? (
          <p className="text-gray-500">Загрузка...</p>
        ) : (
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filtered.map((c) => (
              <div key={c.id} className="bg-white rounded-2xl shadow-sm p-6 border border-gray-100">
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h3 className="font-bold text-lg text-gray-900">{c.name}</h3>
                    <p className="text-gray-500 text-sm">{c.city || "—"}</p>
                  </div>
                  <StatusBadge status={c.status} />
                </div>
                <div className="text-sm text-gray-600 space-y-1">
                  <p>Участников: <span className="font-medium">{c.members || 0}</span></p>
                  {c.university && <p>Вуз: {c.university}</p>}
                </div>
              </div>
            ))}
            {filtered.length === 0 && (
              <p className="text-gray-500 text-center py-12 col-span-full">Клубы не найдены.</p>
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
    archived: "bg-gray-100 text-gray-600",
  };
  return (
    <span className={`text-xs px-2 py-1 rounded-full ${colors[status] || "bg-gray-100 text-gray-600"}`}>
      {status === "active" ? "Активен" : "В архиве"}
    </span>
  );
}
