"use client";

import { useState, useEffect, useRef, useCallback } from "react";

interface Player {
  id: number;
  full_name: string;
  nickname: string;
}

interface PlayerSearchProps {
  value: string;
  onSelect: (playerId: string, displayName: string) => void;
  placeholder?: string;
  label?: string;
}

export default function PlayerSearch({ value, onSelect, placeholder = "Поиск игрока...", label }: PlayerSearchProps) {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<Player[]>([]);
  const [loading, setLoading] = useState(false);
  const [showDropdown, setShowDropdown] = useState(false);
  const [selectedName, setSelectedName] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);
  const debounceRef = useRef<NodeJS.Timeout | null>(null);

  const searchPlayers = useCallback(async (searchQuery: string) => {
    if (searchQuery.length < 2) {
      setResults([]);
      return;
    }
    setLoading(true);
    try {
      const res = await fetch(`/api/v1/players?q=${encodeURIComponent(searchQuery)}`);
      if (res.ok) {
        const data = await res.json();
        setResults(data || []);
      }
    } catch {
      setResults([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }
    debounceRef.current = setTimeout(() => {
      searchPlayers(query);
    }, 300);
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, [query, searchPlayers]);

  function handleSelect(player: Player) {
    const displayName = `${player.full_name} (@${player.nickname})`;
    setSelectedName(displayName);
    setQuery("");
    setResults([]);
    setShowDropdown(false);
    onSelect(String(player.id), displayName);
  }

  return (
    <div className="relative">
      {label && (
        <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
      )}
      <input
        ref={inputRef}
        type="text"
        value={selectedName || query}
        onChange={(e) => {
          if (selectedName) {
            setSelectedName("");
            onSelect("", "");
          }
          setQuery(e.target.value);
          setShowDropdown(true);
        }}
        onFocus={() => setShowDropdown(true)}
        placeholder={placeholder}
        className="w-full px-4 py-3 rounded-lg border border-gray-300 focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all"
      />

      {showDropdown && (query.length >= 2 || results.length > 0) && (
        <div className="absolute z-50 w-full mt-1 bg-white rounded-xl shadow-lg border border-gray-200 max-h-60 overflow-auto">
          {loading ? (
            <div className="px-4 py-3 text-gray-500 text-sm">Поиск...</div>
          ) : results.length === 0 ? (
            <div className="px-4 py-3 text-gray-500 text-sm">Ничего не найдено</div>
          ) : (
            results.map((player) => (
              <button
                key={player.id}
                onClick={() => handleSelect(player)}
                className="w-full text-left px-4 py-3 hover:bg-primary-50 transition-colors flex items-center gap-3"
              >
                <div className="w-8 h-8 rounded-full bg-primary-100 flex items-center justify-center text-primary-700 font-bold text-sm">
                  {player.nickname[0]?.toUpperCase()}
                </div>
                <div>
                  <div className="font-medium text-gray-900">{player.full_name}</div>
                  <div className="text-sm text-gray-500">@{player.nickname}</div>
                </div>
              </button>
            ))
          )}
        </div>
      )}
    </div>
  );
}
