"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { api, Season } from "@/lib/api";
import { adminApi } from "@/lib/admin-api";

export default function AdminSeasonsPage() {
  const router = useRouter();
  const [seasons, setSeasons] = useState<Season[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [showCreate, setShowCreate] = useState(false);

  // Create form
  const [number, setNumber] = useState("");
  const [title, setTitle] = useState("");
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem("admin_token");
    if (!token) {
      router.push("/admin/login");
      return;
    }
    loadSeasons();
  }, [router]);

  function loadSeasons() {
    setLoading(true);
    api.seasons()
      .then(setSeasons)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setCreating(true);
    try {
      await adminApi.createSeason({
        number: Number(number),
        title,
        start_date: startDate,
        end_date: endDate,
      });
      setShowCreate(false);
      setNumber(""); setTitle(""); setStartDate(""); setEndDate("");
      loadSeasons();
    } catch (err: any) {
      setError(err.message);
    } finally {
      setCreating(false);
    }
  }

  async function transition(id: string, action: string, fn: () => Promise<void>) {
    if (!confirm(`Подтвердите: ${action}?`)) return;
    try {
      await fn();
      loadSeasons();
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
          <h1 className="text-3xl font-bold text-gray-900">Управление сезонами</h1>
          <button
            onClick={() => setShowCreate(!showCreate)}
            className="bg-primary-600 text-white px-6 py-2 rounded-lg font-medium hover:bg-primary-700 transition-colors"
          >
            {showCreate ? "Отмена" : "+ Создать сезон"}
          </button>
        </div>

        {error && <p className="text-red-600 mb-4">{error}</p>}

        {showCreate && (
          <form onSubmit={handleCreate} className="bg-white rounded-2xl shadow-lg p-6 mb-8 space-y-4">
            <h2 className="text-lg font-bold text-gray-900">Новый сезон</h2>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Номер *</label>
                <input type="number" value={number} onChange={(e) => setNumber(e.target.value)} required className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none" />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Название *</label>
                <input type="text" value={title} onChange={(e) => setTitle(e.target.value)} required className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none" />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Начало *</label>
                <input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} required className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none" />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Конец *</label>
                <input type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)} required className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none" />
              </div>
            </div>
            <button type="submit" disabled={creating} className="bg-primary-600 text-white px-6 py-2 rounded-lg font-medium hover:bg-primary-700 transition-colors disabled:opacity-50">
              {creating ? "Создание..." : "Создать"}
            </button>
          </form>
        )}

        {loading ? <p className="text-gray-500">Загрузка...</p> : (
          <div className="space-y-4">
            {seasons.map((s) => (
              <SeasonAdminCard key={s.id} season={s} onAction={(action, fn) => transition(s.id, action, fn)} />
            ))}
            {seasons.length === 0 && <p className="text-gray-500 text-center py-12">Пока нет сезонов.</p>}
          </div>
        )}
      </div>
    </main>
  );
}

function SeasonAdminCard({ season, onAction }: { season: Season; onAction: (action: string, fn: () => Promise<void>) => void }) {
  const statusLabels: Record<string, string> = {
    announced: "Анонсирован", registration: "Регистрация", active: "Активен",
    closing: "Закрытие", finished: "Завершён", archived: "В архиве",
  };

  const statusColors: Record<string, string> = {
    announced: "bg-blue-100 text-blue-700", registration: "bg-green-100 text-green-700",
    active: "bg-primary-100 text-primary-700", closing: "bg-yellow-100 text-yellow-700",
    finished: "bg-gray-100 text-gray-600", archived: "bg-gray-100 text-gray-500",
  };

  return (
    <div className="bg-white rounded-xl shadow-sm p-6 border border-gray-100">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="font-bold text-lg text-gray-900">{season.title}</h3>
          <p className="text-gray-500 text-sm">
            {new Date(season.start_date).toLocaleDateString("ru-RU")} — {new Date(season.end_date).toLocaleDateString("ru-RU")}
          </p>
        </div>
        <span className={`text-sm px-3 py-1 rounded-full font-medium ${statusColors[season.status]}`}>
          {statusLabels[season.status] || season.status}
        </span>
      </div>

      <div className="flex flex-wrap gap-2">
        {season.status === "announced" && (
          <ActionBtn label="Начать регистрацию" onClick={() => onAction("Начать регистрацию", () => adminApi.startRegistration(season.id))} />
        )}
        {(season.status === "announced" || season.status === "registration") && (
          <ActionBtn label="Старт сезона" onClick={() => onAction("Старт сезона", () => adminApi.startSeason(season.id))} />
        )}
        {season.status === "active" && (
          <ActionBtn label="Закрыть приём" onClick={() => onAction("Закрыть приём", () => adminApi.closeSeason(season.id))} />
        )}
        {(season.status === "active" || season.status === "closing") && (
          <>
            <ActionBtn label="Завершить" onClick={() => onAction("Завершить сезон", () => adminApi.finishSeason(season.id))} />
          </>
        )}
        {season.status === "finished" && (
          <>
            <ActionBtn label="Расчёт результатов" onClick={() => onAction("Расчёт результатов", () => adminApi.calculateResults(season.id))} />
            <ActionBtn label="Ротация" onClick={() => onAction("Применить ротацию", () => adminApi.applyRotation(season.id))} />
            <ActionBtn label="Бейджи" onClick={() => onAction("Выдать бейджи", () => adminApi.awardSeasonBadges(season.id))} />
            <ActionBtn label="Супертитул" onClick={() => onAction("Выдать супертитул", () => adminApi.awardSuperTitle(season.id))} />
            <ActionBtn label="В архив" onClick={() => onAction("В архив", () => adminApi.archiveSeason(season.id))} />
          </>
        )}
      </div>
    </div>
  );
}

function ActionBtn({ label, onClick }: { label: string; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className="bg-gray-100 hover:bg-gray-200 text-gray-700 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors"
    >
      {label}
    </button>
  );
}
