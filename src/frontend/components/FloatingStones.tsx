"use client";

import { useEffect, useState } from "react";

interface Stone {
  id: number;
  x: number;
  size: number;
  duration: number;
  delay: number;
  type: "black" | "white";
  direction: "up" | "diagonal";
}

function generateStones(count: number): Stone[] {
  return Array.from({ length: count }, (_, i) => ({
    id: i,
    x: Math.random() * 100,
    size: 20 + Math.random() * 40,
    duration: 15 + Math.random() * 20,
    delay: Math.random() * 15,
    type: Math.random() > 0.5 ? "black" : "white",
    direction: Math.random() > 0.7 ? "diagonal" : "up",
  }));
}

function GoStone({ size, type }: { size: number; type: "black" | "white" }) {
  const isBlack = type === "black";
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 100 100"
      className="drop-shadow-lg"
      style={{ filter: "drop-shadow(0 4px 6px rgba(0,0,0,0.3))" }}
    >
      <defs>
        <radialGradient
          id={`stoneGrad-${type}`}
          cx="35%"
          cy="35%"
          r="65%"
        >
          {isBlack ? (
            <>
              <stop offset="0%" stopColor="#4a4a4a" />
              <stop offset="50%" stopColor="#1a1a1a" />
              <stop offset="100%" stopColor="#0a0a0a" />
            </>
          ) : (
            <>
              <stop offset="0%" stopColor="#ffffff" />
              <stop offset="50%" stopColor="#f0f0f0" />
              <stop offset="100%" stopColor="#d0d0d0" />
            </>
          )}
        </radialGradient>
      </defs>
      <circle
        cx="50"
        cy="50"
        r="48"
        fill={`url(#stoneGrad-${type})`}
      />
      {/* Highlight */}
      <ellipse
        cx="38"
        cy="32"
        rx="15"
        ry="10"
        fill="rgba(255,255,255,0.15)"
      />
    </svg>
  );
}

export default function FloatingStones() {
  const [stones, setStones] = useState<Stone[]>([]);

  useEffect(() => {
    setStones(generateStones(25));
  }, []);

  return (
    <div className="fixed inset-0 pointer-events-none overflow-hidden z-0">
      {stones.map((stone) => (
        <div
          key={stone.id}
          className={`absolute ${stone.direction === "up" ? "animate-float" : "animate-float-diagonal"}`}
          style={{
            left: `${stone.x}%`,
            animationDuration: `${stone.duration}s`,
            animationDelay: `${stone.delay}s`,
          }}
        >
          <GoStone size={stone.size} type={stone.type} />
        </div>
      ))}
    </div>
  );
}
