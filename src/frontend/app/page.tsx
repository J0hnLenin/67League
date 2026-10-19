"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import Header from "@/components/Header";

export default function HomePage() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem("token");
    setIsLoggedIn(!!token);
  }, []);

  return (
    <main className="min-h-screen bg-gradient-to-b from-slate-50 to-slate-100">
      <Header />

      {/* Hero */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-20 text-center">
        <h1 className="text-5xl sm:text-6xl font-extrabold text-gray-900 mb-6">
          Играй в Го <span className="text-primary-600">с лучшими</span>
        </h1>
        <p className="text-xl text-gray-600 max-w-2xl mx-auto mb-10">
          Студенческая сезонная онлайн-лига с дивизионами, клубами, бейджами и внутриигровой валютой.
          Не pay-to-win — только соревнование и сообщество.
        </p>
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          {isLoggedIn ? (
            <Link
              href="/dashboard"
              className="bg-primary-600 text-white px-8 py-4 rounded-xl text-lg font-semibold hover:bg-primary-700 transition-colors shadow-lg"
            >
              В личный кабинет →
            </Link>
          ) : (
            <>
              <Link
                href="/register"
                className="bg-primary-600 text-white px-8 py-4 rounded-xl text-lg font-semibold hover:bg-primary-700 transition-colors shadow-lg"
              >
                Присоединиться
              </Link>
              <Link
                href="/about"
                className="bg-white text-gray-700 border border-gray-300 px-8 py-4 rounded-xl text-lg font-semibold hover:bg-gray-50 transition-colors"
              >
                Узнать больше
              </Link>
            </>
          )}
        </div>
      </section>

      {/* Features */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
        <div className="grid md:grid-cols-3 gap-8">
          <FeatureCard
            icon="🏆"
            title="Дивизионы"
            description="Соревнуйся в дивизионе S, A или B по своему уровню. Прозрачная ротация каждый сезон."
          />
          <FeatureCard
            icon="🏛️"
            title="Клубы"
            description="Создай или вступи в клуб своего вуза или города. Клубный зачёт и бейджи."
          />
          <FeatureCard
            icon="🎖️"
            title="Бейджи и титулы"
            description="Зарабатывай достижения за победы, стрики и активность. Выбирай титул для ника."
          />
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gray-900 text-gray-400 py-12 mt-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <p className="text-lg font-semibold text-white mb-2">6-7 Лига</p>
          <p className="text-sm">Студенческая лига по Го © 2026</p>
        </div>
      </footer>
    </main>
  );
}

function FeatureCard({ icon, title, description }: { icon: string; title: string; description: string }) {
  return (
    <div className="bg-white rounded-2xl p-8 shadow-sm border border-gray-100 hover:shadow-md transition-shadow">
      <div className="text-4xl mb-4">{icon}</div>
      <h3 className="text-xl font-bold text-gray-900 mb-3">{title}</h3>
      <p className="text-gray-600 leading-relaxed">{description}</p>
    </div>
  );
}
