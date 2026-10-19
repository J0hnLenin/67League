import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "6-7 Лига — Студенческая лига по Го",
  description: "Сезонная онлайн-лига по игре Го для студентов, преподавателей и выпускников",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ru">
      <body className="antialiased min-h-screen">
        {children}
      </body>
    </html>
  );
}
