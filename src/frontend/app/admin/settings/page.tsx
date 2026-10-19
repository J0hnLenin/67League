"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { adminApi } from "@/lib/admin-api";

export default function AdminSettingsPage() {
  const router = useRouter();
  const [settings, setSettings] = useState<Record<string, any>>({});
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  useEffect(() => {
    const token = localStorage.getItem("admin_token");
    if (!token) {
      router.push("/admin/login");
      return;
    }
    loadSettings();
  }, [router]);

  function loadSettings() {
    setLoading(true);
    adminApi
      .getSettings()
      .then((data) => setSettings(data))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }

  async function handleSave() {
    setSaving(true);
    setError("");
    setSuccess("");
    try {
      await adminApi.patchSettings(settings);
      setSuccess("Настройки сохранены");
    } catch (err: any) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  }

  function updateValue(key: string, value: any) {
    setSettings((prev) => ({ ...prev, [key]: value }));
  }

  const settingGroups = [
    {
      title: "Ротация",
      keys: ["k1", "k2", "k3"],
      labels: { k1: "S ↔ A", k2: "A ↔ B", k3: "B ↔ C" },
    },
    {
      title: "Партии",
      keys: ["p", "n", "t"],
      labels: { p: "Мин. партий", n: "Макс. с одним", t: "Порог бонуса" },
    },
    {
      title: "Сроки (дни)",
      keys: ["challenge_accept_days", "match_cancel_days", "season_duration_days", "closing_days"],
      labels: {
        challenge_accept_days: "На принятие",
        match_cancel_days: "На отмену",
        season_duration_days: "Длительность сезона",
        closing_days: "До закрытия",
      },
    },
    {
      title: "LP",
      keys: ["lp_win", "lp_loss"],
      labels: { lp_win: "За победу", lp_loss: "За поражение" },
    },
    {
      title: "Стрик-бонусы",
      keys: ["streak_bonus_3", "streak_bonus_5", "streak_bonus_67"],
      labels: { streak_bonus_3: "3 победы", streak_bonus_5: "5 побед", streak_bonus_67: "6/7 побед" },
    },
    {
      title: "Камни",
      keys: ["stones_win", "stones_loss", "stones_club_top1", "stones_super_title"],
      labels: {
        stones_win: "За победу",
        stones_loss: "За поражение",
        stones_club_top1: "Клуб топ-1",
        stones_super_title: "Супертитул",
      },
    },
  ];

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

      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">Настройки лиги</h1>

        {error && <p className="text-red-600 mb-4">{error}</p>}
        {success && <p className="text-green-600 mb-4">{success}</p>}

        {loading ? (
          <p className="text-gray-500">Загрузка...</p>
        ) : (
          <>
            <div className="space-y-8">
              {settingGroups.map((group) => (
                <div key={group.title} className="bg-white rounded-2xl shadow-sm p-6">
                  <h2 className="text-lg font-bold text-gray-900 mb-4">{group.title}</h2>
                  <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                    {group.keys.map((key) => (
                      <div key={key}>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          {group.labels[key as keyof typeof group.labels] || key}
                        </label>
                        <input
                          type="number"
                          value={settings[key] ?? ""}
                          onChange={(e) => updateValue(key, Number(e.target.value))}
                          className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 outline-none"
                        />
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </div>

            <div className="mt-8 flex justify-end">
              <button
                onClick={handleSave}
                disabled={saving}
                className="bg-primary-600 text-white px-8 py-3 rounded-lg font-medium hover:bg-primary-700 transition-colors disabled:opacity-50"
              >
                {saving ? "Сохранение..." : "Сохранить"}
              </button>
            </div>
          </>
        )}
      </div>
    </main>
  );
}
