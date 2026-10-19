"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";

export default function RegisterPage() {
  const router = useRouter();
  const [form, setForm] = useState({
    full_name: "",
    email: "",
    password: "",
    nickname: "",
    city: "",
    kgs_nick: "",
    ogs_nick: "",
    fox_nick: "",
    rfg_id: "",
    sputnik_id: "",
  });
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  function updateField(field: string, value: string) {
    setForm((prev) => ({ ...prev, [field]: value }));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const res = await api.register(form);
      localStorage.setItem("token", res.token);
      router.push("/dashboard");
    } catch (err: any) {
      setError(err.message || "Ошибка регистрации");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen flex items-center justify-center bg-gradient-to-b from-slate-50 to-slate-100 px-4 py-12">
      <div className="w-full max-w-lg">
        <div className="text-center mb-8">
          <Link href="/" className="text-3xl font-bold text-primary-700">
            6-7 Лига
          </Link>
          <p className="text-gray-500 mt-2">Регистрация</p>
        </div>

        <form onSubmit={handleSubmit} className="bg-white rounded-2xl shadow-lg p-8 space-y-5">
          {error && (
            <div className="bg-red-50 text-red-700 px-4 py-3 rounded-lg text-sm">
              {error}
            </div>
          )}

          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Имя и фамилия *"
              value={form.full_name}
              onChange={(v) => updateField("full_name", v)}
              required
            />
            <Input
              label="Ник в лиге *"
              value={form.nickname}
              onChange={(v) => updateField("nickname", v)}
              required
            />
          </div>

          <Input
            label="Email *"
            type="email"
            value={form.email}
            onChange={(v) => updateField("email", v)}
            required
          />

          <Input
            label="Пароль *"
            type="password"
            value={form.password}
            onChange={(v) => updateField("password", v)}
            required
            minLength={6}
          />

          <Input
            label="Город *"
            value={form.city}
            onChange={(v) => updateField("city", v)}
            required
          />

          <div className="border-t pt-4">
            <p className="text-sm text-gray-500 mb-3">Ники на серверах (опционально)</p>
            <div className="grid grid-cols-2 gap-4">
              <Input
                label="KGS"
                value={form.kgs_nick}
                onChange={(v) => updateField("kgs_nick", v)}
              />
              <Input
                label="OGS"
                value={form.ogs_nick}
                onChange={(v) => updateField("ogs_nick", v)}
              />
              <Input
                label="Fox"
                value={form.fox_nick}
                onChange={(v) => updateField("fox_nick", v)}
              />
            </div>
          </div>

          <div className="border-t pt-4">
            <p className="text-sm text-gray-500 mb-3">Рейтинг (опционально)</p>
            <div className="grid grid-cols-2 gap-4">
              <Input
                label="ID на РФГ"
                value={form.rfg_id}
                onChange={(v) => updateField("rfg_id", v)}
              />
              <Input
                label="ID на Спутник"
                value={form.sputnik_id}
                onChange={(v) => updateField("sputnik_id", v)}
              />
            </div>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-primary-600 text-white py-3 rounded-lg font-semibold hover:bg-primary-700 transition-colors disabled:opacity-50"
          >
            {loading ? "Регистрация..." : "Зарегистрироваться"}
          </button>
        </form>

        <p className="text-center mt-6 text-gray-600">
          Уже есть аккаунт?{" "}
          <Link href="/login" className="text-primary-600 hover:underline font-medium">
            Войти
          </Link>
        </p>
      </div>
    </main>
  );
}

function Input({
  label,
  value,
  onChange,
  type = "text",
  required,
  minLength,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  type?: string;
  required?: boolean;
  minLength?: number;
}) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
      <input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        required={required}
        minLength={minLength}
        className="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
      />
    </div>
  );
}
