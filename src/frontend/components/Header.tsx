"use client";

import { useEffect, useState } from "react";
import Link from "next/link";

export default function Header() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem("token");
    setIsLoggedIn(!!token);
  }, []);

  return (
    <header className="bg-white shadow-sm">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        <Link href="/" className="flex items-center gap-2">
          <span className="text-2xl font-bold text-primary-700">6-7 Лига</span>
          <span className="text-sm text-gray-500 hidden sm:inline">Студенческая лига по Го</span>
        </Link>
        <nav className="flex items-center gap-4">
          <Link href="/clubs" className="text-gray-600 hover:text-primary-600 transition-colors">
            Клубы
          </Link>
          <Link href="/seasons" className="text-gray-600 hover:text-primary-600 transition-colors">
            Сезоны
          </Link>
          {isLoggedIn ? (
            <Link
              href="/dashboard"
              className="bg-primary-600 text-white px-4 py-2 rounded-lg hover:bg-primary-700 transition-colors"
            >
              Личный кабинет
            </Link>
          ) : (
            <>
              <Link href="/login" className="text-gray-600 hover:text-primary-600 transition-colors">
                Войти
              </Link>
              <Link
                href="/register"
                className="bg-primary-600 text-white px-4 py-2 rounded-lg hover:bg-primary-700 transition-colors"
              >
                Регистрация
              </Link>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
