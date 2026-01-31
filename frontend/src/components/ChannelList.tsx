import { useState } from "react";
import type { FormEvent } from "react";
import type { Channel } from "../types";

interface ChannelListProps {
  channels: Channel[];
  currentChannelId: number | null;
  onSelect: (id: number) => void;
  onCreate: (name: string, description: string) => Promise<void>;
}

export default function ChannelList({
  channels,
  currentChannelId,
  onSelect,
  onCreate,
}: ChannelListProps) {
  const [showForm, setShowForm] = useState(false);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [creating, setCreating] = useState(false);

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    const trimmed = name.trim();
    if (!trimmed || creating) return;
    setCreating(true);
    try {
      await onCreate(trimmed, description.trim());
      setName("");
      setDescription("");
      setShowForm(false);
    } finally {
      setCreating(false);
    }
  }

  return (
    <div>
      {/* Section header */}
      <div className="flex items-center justify-between px-4 mb-1">
        <p className="text-[#a0a0b8]/40 text-[10px] uppercase tracking-widest">
          Channels
        </p>
        <button
          onClick={() => setShowForm(!showForm)}
          className="text-[#a0a0b8]/40 hover:text-[#e2b714] text-lg leading-none transition-colors cursor-pointer"
          title="New channel"
        >
          {showForm ? "\u00D7" : "+"}
        </button>
      </div>

      {/* Inline create form */}
      {showForm && (
        <form onSubmit={handleCreate} className="px-3 mb-2">
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="channel-name"
            disabled={creating}
            className="w-full bg-[#0f3460]/50 border border-[#e2b714]/10 rounded px-2 py-1.5 text-white text-xs placeholder-[#a0a0b8]/30 focus:outline-none focus:border-[#e2b714]/30 font-mono disabled:opacity-50"
          />
          <input
            type="text"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="description (optional)"
            disabled={creating}
            className="w-full bg-[#0f3460]/50 border border-[#e2b714]/10 rounded px-2 py-1.5 text-white text-xs placeholder-[#a0a0b8]/30 focus:outline-none focus:border-[#e2b714]/30 font-mono mt-1 disabled:opacity-50"
          />
          <button
            type="submit"
            disabled={!name.trim() || creating}
            className="mt-1.5 w-full bg-[#e2b714]/15 hover:bg-[#e2b714]/25 disabled:bg-[#e2b714]/5 text-[#e2b714] disabled:text-[#e2b714]/30 text-xs py-1.5 rounded transition-colors cursor-pointer disabled:cursor-not-allowed"
          >
            {creating ? "creating..." : "Create"}
          </button>
        </form>
      )}

      {/* Channel list */}
      <div className="space-y-px">
        {channels.map((ch) => {
          const active = ch.id === currentChannelId;
          return (
            <button
              key={ch.id}
              onClick={() => onSelect(ch.id)}
              className={`w-full text-left px-4 py-1.5 text-sm flex items-center transition-colors cursor-pointer ${
                active
                  ? "text-[#e2b714] bg-[#e2b714]/8 border-l-2 border-[#e2b714]"
                  : "text-[#a0a0b8]/70 hover:text-[#a0a0b8] hover:bg-white/3 border-l-2 border-transparent"
              }`}
            >
              <span
                className={`mr-1.5 text-xs ${active ? "text-[#e2b714]/60" : "text-[#a0a0b8]/30"}`}
              >
                #
              </span>
              <span className="truncate">{ch.name}</span>
            </button>
          );
        })}
      </div>

      {channels.length === 0 && !showForm && (
        <p className="px-4 text-[#a0a0b8]/30 text-xs italic">no channels yet</p>
      )}
    </div>
  );
}
