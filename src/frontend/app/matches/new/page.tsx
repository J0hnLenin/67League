"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";

export default function NewMatchPage() {
  const router = useRouter();
  const [form, setForm] = useState({
    season_id: "",
    division_id: "",
    black_player_id: "",
    white_player_id: "",
    server: "ogs" as "ogs" | "kgs" | "fox" | "offline",
    time_control: "",
    handicap: 0,
    komi: 6.5,
  });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  function updateField(field: string, value: string | number) {
    setForm((prev) => ({ ...prev, [field]: value }));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      await api.createMatch(form);
      router.push("/matches");
    } catch (err: any) {
      setError(err.message || "Ошибка создания вызова");
    } finally {
      setLoading(false);
    }
  }

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

      <div className="max-w-2xl mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">Создать вызов</h1>

        <form onSubmit={handleSubmit} className="bg-white rounded-2xl shadow-lg p-8 space-y-5">
          {error && (
            <div className="bg-red-50 text-red-700 px-4 py-3 rounded-lg text-sm">{error}</div>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Сезон ID *</label>
              <input
                type="text"
                value={form.season_id}
                onChange={(e) => updateField("season_id", e.target.value)}
                required
                className="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Дивизион ID *</label>
              <input
                type="text"
                value={form.division_id}
                onChange={(e) => updateField("division_id", e.target.value)}
                required
                className="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Чёрные (ID) *</label>
              <input
                type="text"
                value={form.black_player_id}
                onChange={(e) => updateField("black_player_id", e.target.value)}
                required
                className="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Белые (ID) *</label>
              <input
                type="text"
                value={form.white_player_id}
                onChange={(e) => updateField("white_player_id", e.target.value)}
                required
                className="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Сервер *</label>
            <select
              value={form.server}
              onChange={(e) => updateField("server", e.target.value)}
              className="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none"
            >
              <option value="ogs">OGS</option>
              <option value="kgs">KGS</option>
              <option value="fox">Fox</option>
              <option value="offline">Оффлайн</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Контроль времени</label>
            <input
              type="text"
              value={form.time_control}
              onChange={(e) => updateField("time_control", e.target.value)}
              placeholder="byoyomi:30:3x30"
              className="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Фора</label>
              <input
                type="number"
                value={form.handicap}
                onChange={(e) => updateField("handicap", Number(e.target.value))}
                className="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Коми</label>
              <input
                type="number"
                step="0.5"
                value={form.komi}
                onChange={(e) => updateField("komi", Number(e.target.value))}
                className="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none"
              />
            </div>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-primary-600 text-white py-3 rounded-lg font-semibold hover:bg-primary-700 transition-colors disabled:opacity-50"
          >
            {loading ? "Создание..." : "Создать вызов"}
          </button>
        </form>
      </div>
    </main>
  );
}
