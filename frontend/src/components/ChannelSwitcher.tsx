import { useState, useEffect, useRef } from "react";
import type { Channel } from "../types";

interface ChannelSwitcherProps {
  channels: Channel[];
  onSelect: (id: number) => void;
  onClose: () => void;
}

export default function ChannelSwitcher({
  channels,
  onSelect,
  onClose,
}: ChannelSwitcherProps) {
  const [query, setQuery] = useState("");
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  const filtered = channels.filter((ch) =>
    ch.name.toLowerCase().includes(query.toLowerCase()),
  );

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  useEffect(() => {
    setSelectedIndex(0);
  }, [query]);

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Escape") {
      onClose();
    } else if (e.key === "ArrowDown") {
      e.preventDefault();
      setSelectedIndex((i) => Math.min(i + 1, filtered.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setSelectedIndex((i) => Math.max(i - 1, 0));
    } else if (e.key === "Enter" && filtered[selectedIndex]) {
      onSelect(filtered[selectedIndex].id);
      onClose();
    }
  }

  return (
    <div
      className="fixed inset-0 z-40 flex items-start justify-center pt-[20vh] bg-black/50 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="bg-[#16213e] border border-[#e2b714]/20 rounded-lg shadow-2xl w-full max-w-md mx-4 overflow-hidden animate-[slideDown_0.15s_ease-out]"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="px-4 py-3 border-b border-[#e2b714]/10">
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Switch channel..."
            className="w-full bg-transparent text-white text-sm font-mono placeholder:text-[#a0a0b8]/30 focus:outline-none"
          />
        </div>
        <div className="max-h-64 overflow-y-auto py-1">
          {filtered.length === 0 ? (
            <div className="px-4 py-3 text-[#a0a0b8]/40 text-sm font-mono text-center">
              no matching channels
            </div>
          ) : (
            filtered.map((ch, i) => (
              <button
                key={ch.id}
                onClick={() => {
                  onSelect(ch.id);
                  onClose();
                }}
                className={`w-full text-left px-4 py-2 text-sm font-mono flex items-center gap-2 transition-colors cursor-pointer ${
                  i === selectedIndex
                    ? "bg-[#e2b714]/10 text-[#e2b714]"
                    : "text-[#a0a0b8]/70 hover:bg-white/3"
                }`}
              >
                <span className="text-[#a0a0b8]/30 text-xs">#</span>
                <span className="truncate">{ch.name}</span>
                {ch.description && (
                  <span className="text-[#a0a0b8]/25 text-xs truncate ml-auto">
                    {ch.description}
                  </span>
                )}
              </button>
            ))
          )}
        </div>
        <div className="px-4 py-2 border-t border-[#e2b714]/10 flex items-center gap-3 text-[10px] text-[#a0a0b8]/30 font-mono">
          <span>
            <kbd className="bg-[#0f3460]/50 px-1 py-0.5 rounded text-[#a0a0b8]/50">
              ↑↓
            </kbd>{" "}
            navigate
          </span>
          <span>
            <kbd className="bg-[#0f3460]/50 px-1 py-0.5 rounded text-[#a0a0b8]/50">
              ↵
            </kbd>{" "}
            select
          </span>
          <span>
            <kbd className="bg-[#0f3460]/50 px-1 py-0.5 rounded text-[#a0a0b8]/50">
              esc
            </kbd>{" "}
            close
          </span>
        </div>
      </div>
    </div>
  );
}
