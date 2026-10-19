"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import Header from "@/components/Header";
import { api, Player } from "@/lib/api";

export default function DashboardPage() {
  const [player, setPlayer] = useState<Player | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    api.me()
      .then(setPlayer)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <main className="min-h-screen flex items-center justify-center">
        <div className="text-gray-500">Загрузка...</div>
      </main>
    );
  }

  if (error || !player) {
    return (
      <main className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 mb-4">{error || "Не удалось загрузить профиль"}</p>
          <Link href="/login" className="text-primary-600 hover:underline">
            Войти снова
          </Link>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-gray-50">
      <Header />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid lg:grid-cols-3 gap-8">
          {/* Profile Card */}
          <div className="lg:col-span-1">
            <div className="bg-white rounded-2xl shadow-sm p-6">
              <div className="text-center">
                <div className="w-20 h-20 bg-primary-100 rounded-full flex items-center justify-center mx-auto mb-4">
                  <span className="text-3xl">👤</span>
                </div>
                <h2 className="text-xl font-bold text-gray-900">{player.full_name}</h2>
                <p className="text-gray-500">@{player.nickname}</p>
                <p className="text-gray-500 text-sm">{player.city}</p>
              </div>

              <div className="mt-6 space-y-3">
                <StatRow label="Рейтинг РФГ" value={player.rating || "—"} />
                <StatRow label="Камни" value={player.stones} />
                <StatRow label="LP (всего)" value={player.lp_total} />
                <StatRow label="Статус" value={statusLabel(player.status)} />
              </div>
            </div>
          </div>

          {/* Main Content */}
          <div className="lg:col-span-2 space-y-6">
            {/* Quick Actions */}
            <div className="bg-white rounded-2xl shadow-sm p-6">
              <h3 className="text-lg font-bold text-gray-900 mb-4">Быстрые действия</h3>
              <div className="grid sm:grid-cols-3 gap-4">
                <ActionCard href="/matches/new" icon="⚔️" title="Создать вызов" />
                <ActionCard href="/clubs" icon="🏛️" title="Клубы" />
                <ActionCard href="/seasons" icon="📅" title="Сезоны" />
              </div>
            </div>

            {/* Links */}
            <div className="bg-white rounded-2xl shadow-sm p-6">
              <h3 className="text-lg font-bold text-gray-900 mb-4">Моё</h3>
              <div className="space-y-2">
                <LinkRow href="/matches" label="Мои партии" />
                <LinkRow href="/badges" label="Мои бейджи" />
                <LinkRow href="/stones" label="История камней" />
                <LinkRow href="/settings" label="Настройки" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}

function StatRow({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="flex justify-between items-center py-2 border-b border-gray-100 last:border-0">
      <span className="text-gray-600 text-sm">{label}</span>
      <span className="font-semibold text-gray-900">{value}</span>
    </div>
  );
}

function ActionCard({ href, icon, title }: { href: string; icon: string; title: string }) {
  return (
    <Link
      href={href}
      className="flex flex-col items-center p-4 rounded-xl border border-gray-200 hover:border-primary-300 hover:bg-primary-50 transition-colors"
    >
      <span className="text-3xl mb-2">{icon}</span>
      <span className="text-sm font-medium text-gray-700">{title}</span>
    </Link>
  );
}

function LinkRow({ href, label }: { href: string; label: string }) {
  return (
    <Link
      href={href}
      className="flex items-center justify-between py-3 px-4 rounded-lg hover:bg-gray-50 transition-colors"
    >
      <span className="text-gray-700">{label}</span>
      <span className="text-gray-400">→</span>
    </Link>
  );
}

function statusLabel(status: string): string {
  switch (status) {
    case "active":
      return "Активен";
    case "inactive":
      return "Неактивен";
    case "banned":
      return "Забанен";
    default:
      return status;
  }
}
