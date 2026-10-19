"use client";

import { useEffect, useState, useRef } from "react";
import Link from "next/link";
import Header from "@/components/Header";
import FloatingStones from "@/components/FloatingStones";

export default function HomePage() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [counters, setCounters] = useState({ players: 0, matches: 0, clubs: 0 });
  const statsRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const token = localStorage.getItem("token");
    setIsLoggedIn(!!token);

    // Animate counters when visible
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          animateCounter(128, "players");
          animateCounter(567, "matches");
          animateCounter(24, "clubs");
          observer.disconnect();
        }
      },
      { threshold: 0.5 }
    );

    if (statsRef.current) {
      observer.observe(statsRef.current);
    }

    return () => observer.disconnect();
  }, []);

  function animateCounter(target: number, key: keyof typeof counters) {
    const duration = 2000;
    const steps = 60;
    const increment = target / steps;
    let current = 0;

    const timer = setInterval(() => {
      current += increment;
      if (current >= target) {
        current = target;
        clearInterval(timer);
      }
      setCounters((prev) => ({ ...prev, [key]: Math.floor(current) }));
    }, duration / steps);
  }

  return (
    <main className="min-h-screen relative overflow-hidden">
      {/* Animated gradient background */}
      <div className="fixed inset-0 bg-gradient-to-br from-slate-900 via-primary-900 to-slate-900 animate-gradient z-0" />

      {/* Floating stones */}
      <FloatingStones />

      {/* Grid pattern overlay */}
      <div
        className="fixed inset-0 opacity-10 z-0"
        style={{
          backgroundImage: `linear-gradient(rgba(255,255,255,0.1) 1px, transparent 1px),
                           linear-gradient(90deg, rgba(255,255,255,0.1) 1px, transparent 1px)`,
          backgroundSize: "60px 60px",
        }}
      />

      <div className="relative z-10">
        <Header />

        {/* Hero Section */}
        <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-24 pb-20 text-center">
          <div className="animate-fade-in-up">
            <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full glass mb-8">
              <span className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
              <span className="text-white/80 text-sm">Сезон 1 — регистрация открыта</span>
            </div>

            <h1 className="text-6xl sm:text-7xl lg:text-8xl font-black text-white mb-6 tracking-tight">
              6-7 <span className="text-transparent bg-clip-text bg-gradient-to-r from-primary-400 to-cyan-300">Лига</span>
            </h1>

            <p className="text-2xl sm:text-3xl text-white/90 font-light mb-4">
              Студенческая лига по Го
            </p>

            <p className="text-lg text-white/60 max-w-2xl mx-auto mb-12 leading-relaxed">
              Сезонная онлайн-лига с дивизионами, клубами, бейджами и внутриигровой валютой.
              <br />
              Не pay-to-win — только соревнование и сообщество.
            </p>

            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              {isLoggedIn ? (
                <Link
                  href="/dashboard"
                  className="group relative px-8 py-4 bg-white text-primary-900 rounded-2xl text-lg font-bold hover:scale-105 transition-transform animate-pulse-glow"
                >
                  <span className="relative z-10">В личный кабинет →</span>
                  <div className="absolute inset-0 rounded-2xl bg-gradient-to-r from-primary-400 to-cyan-400 opacity-0 group-hover:opacity-20 transition-opacity" />
                </Link>
              ) : (
                <>
                  <Link
                    href="/register"
                    className="group relative px-8 py-4 bg-white text-primary-900 rounded-2xl text-lg font-bold hover:scale-105 transition-transform animate-pulse-glow"
                  >
                    <span className="relative z-10">Присоединиться</span>
                  </Link>
                  <Link
                    href="/login"
                    className="px-8 py-4 glass text-white rounded-2xl text-lg font-semibold hover:bg-white/20 transition-colors"
                  >
                    Уже есть аккаунт
                  </Link>
                </>
              )}
            </div>
          </div>
        </section>

        {/* Stats Section */}
        <section ref={statsRef} className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
          <div className="grid grid-cols-3 gap-8">
            <StatCard value={counters.players} label="Игроков" icon="👤" />
            <StatCard value={counters.matches} label="Партий" icon="⚔️" />
            <StatCard value={counters.clubs} label="Клубов" icon="🏛️" />
          </div>
        </section>

        {/* Features Section */}
        <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-20">
          <h2 className="text-4xl font-bold text-white text-center mb-16">
            Как это <span className="text-primary-400">работает</span>
          </h2>
          <div className="grid md:grid-cols-3 gap-8">
            <FeatureCard
              icon="🏆"
              title="Дивизионы"
              description="Соревнуйся в дивизионе S, A или B по своему уровню. Прозрачная ротация каждый сезон."
              delay={0}
            />
            <FeatureCard
              icon="🏛️"
              title="Клубы"
              description="Создай или вступи в клуб своего вуза или города. Клубный зачёт и бейджи."
              delay={0.2}
            />
            <FeatureCard
              icon="🎖️"
              title="Бейджи и титулы"
              description="Зарабатывай достижения за победы, стрики и активность. Выбирай титул для ника."
              delay={0.4}
            />
          </div>
        </section>

        {/* How it works steps */}
        <section className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-20">
          <div className="glass-dark rounded-3xl p-8 sm:p-12">
            <h2 className="text-3xl font-bold text-white text-center mb-12">
              Начни за <span className="text-primary-400">3 шага</span>
            </h2>
            <div className="grid sm:grid-cols-3 gap-8">
              <StepCard
                step="1"
                title="Регистрация"
                description="Создай аккаунт, укажи свой рейтинг и ник на серверах"
              />
              <StepCard
                step="2"
                title="Вступи в сезон"
                description="Попади в дивизион по своему уровню и найди соперников"
              />
              <StepCard
                step="3"
                title="Играй и побеждай"
                description="Создавай вызовы, играй партии, зарабатывай LP и камни"
              />
            </div>
          </div>
        </section>

        {/* CTA Section */}
        <section className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-20 text-center">
          <div className="glass rounded-3xl p-12">
            <h2 className="text-4xl font-bold text-white mb-4">
              Готов к игре?
            </h2>
            <p className="text-white/70 text-lg mb-8">
              Присоединяйся к сообществу студентов, преподавателей и выпускников,
              которые играют в Го на профессиональном уровне.
            </p>
            <Link
              href={isLoggedIn ? "/dashboard" : "/register"}
              className="inline-block px-10 py-5 bg-gradient-to-r from-primary-500 to-cyan-500 text-white rounded-2xl text-xl font-bold hover:scale-105 transition-transform shadow-lg shadow-primary-500/30"
            >
              {isLoggedIn ? "В личный кабинет →" : "Начать путь →"}
            </Link>
          </div>
        </section>

        {/* Footer */}
        <footer className="border-t border-white/10 py-12 mt-20">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
            <p className="text-2xl font-bold text-white mb-2">6-7 Лига</p>
            <p className="text-white/40 text-sm">Студенческая лига по Го © 2026</p>
          </div>
        </footer>
      </div>
    </main>
  );
}

function StatCard({ value, label, icon }: { value: number; label: string; icon: string }) {
  return (
    <div className="glass rounded-2xl p-6 text-center hover:bg-white/20 transition-colors">
      <div className="text-3xl mb-2">{icon}</div>
      <div className="text-4xl font-black text-white mb-1 animate-count-up">{value}</div>
      <div className="text-white/60 text-sm">{label}</div>
    </div>
  );
}

function FeatureCard({ icon, title, description, delay }: { icon: string; title: string; description: string; delay: number }) {
  const ref = useRef<HTMLDivElement>(null);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setTimeout(() => setVisible(true), delay * 1000);
          observer.disconnect();
        }
      },
      { threshold: 0.3 }
    );
    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, [delay]);

  return (
    <div
      ref={ref}
      className={`glass rounded-2xl p-8 hover:bg-white/20 transition-all duration-500 hover:-translate-y-2 ${
        visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-8"
      }`}
    >
      <div className="text-5xl mb-4">{icon}</div>
      <h3 className="text-xl font-bold text-white mb-3">{title}</h3>
      <p className="text-white/60 leading-relaxed">{description}</p>
    </div>
  );
}

function StepCard({ step, title, description }: { step: string; title: string; description: string }) {
  return (
    <div className="text-center">
      <div className="w-16 h-16 rounded-full bg-gradient-to-br from-primary-500 to-cyan-500 flex items-center justify-center mx-auto mb-4 text-2xl font-black text-white">
        {step}
      </div>
      <h3 className="text-lg font-bold text-white mb-2">{title}</h3>
      <p className="text-white/50 text-sm">{description}</p>
    </div>
  );
}
