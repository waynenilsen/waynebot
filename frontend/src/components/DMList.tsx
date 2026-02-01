import { useState, useEffect } from "react";
import type { DMChannel, Persona } from "../types";
import * as api from "../api";

interface DMListProps {
  dms: DMChannel[];
  currentChannelId: number | null;
  onSelect: (id: number) => void;
  onCreate: (opts: {
    user_id?: number;
    persona_id?: number;
  }) => Promise<void>;
}

function dmDisplayName(dm: DMChannel): string {
  return (
    dm.other_participant.user_name ??
    dm.other_participant.persona_name ??
    dm.name
  );
}

function isPersonaDM(dm: DMChannel): boolean {
  return dm.other_participant.persona_id !== null;
}

export default function DMList({
  dms,
  currentChannelId,
  onSelect,
  onCreate,
}: DMListProps) {
  const [showPicker, setShowPicker] = useState(false);
  const [personas, setPersonas] = useState<Persona[]>([]);
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    if (!showPicker) return;
    api.getPersonas().then(setPersonas).catch(() => {});
  }, [showPicker]);

  async function handleCreate(opts: {
    user_id?: number;
    persona_id?: number;
  }) {
    if (creating) return;
    setCreating(true);
    try {
      await onCreate(opts);
      setShowPicker(false);
    } finally {
      setCreating(false);
    }
  }

  return (
    <div>
      {/* Section header */}
      <div className="flex items-center justify-between px-4 mb-1">
        <p className="text-[#a0a0b8]/40 text-[10px] uppercase tracking-widest">
          Direct Messages
        </p>
        <button
          onClick={() => setShowPicker(!showPicker)}
          className="text-[#a0a0b8]/40 hover:text-[#e2b714] text-lg leading-none transition-colors cursor-pointer"
          title="New DM"
        >
          {showPicker ? "\u00D7" : "+"}
        </button>
      </div>

      {/* Persona picker */}
      {showPicker && (
        <div className="px-3 mb-2">
          {personas.length === 0 && (
            <p className="text-[#a0a0b8]/30 text-xs italic px-1 py-1">
              loading...
            </p>
          )}
          {personas.map((p) => (
            <button
              key={`p-${p.id}`}
              onClick={() => handleCreate({ persona_id: p.id })}
              disabled={creating}
              className="w-full text-left px-2 py-1.5 rounded text-sm flex items-center gap-2 transition-colors cursor-pointer text-[#a0a0b8]/70 hover:text-[#a0a0b8] hover:bg-white/3 disabled:opacity-50"
            >
              <span className="text-xs w-4 text-center text-[#a0a0b8]/30">
                ◆
              </span>
              <span className="truncate">{p.name}</span>
              <span className="text-[#a0a0b8]/25 text-[10px] ml-auto">
                persona
              </span>
            </button>
          ))}
        </div>
      )}

      {/* DM list */}
      <div className="space-y-px">
        {dms.map((dm) => {
          const active = dm.id === currentChannelId;
          const unread = dm.unread_count ?? 0;
          const hasUnread = unread > 0 && !active;
          const persona = isPersonaDM(dm);
          return (
            <button
              key={dm.id}
              onClick={() => onSelect(dm.id)}
              className={`w-full text-left px-4 py-1.5 text-sm flex items-center transition-colors cursor-pointer ${
                active
                  ? "text-[#e2b714] bg-[#e2b714]/8 border-l-2 border-[#e2b714]"
                  : hasUnread
                    ? "text-white hover:bg-white/3 border-l-2 border-transparent font-semibold"
                    : "text-[#a0a0b8]/70 hover:text-[#a0a0b8] hover:bg-white/3 border-l-2 border-transparent"
              }`}
            >
              <span
                className={`mr-1.5 text-xs ${active ? "text-[#e2b714]/60" : "text-[#a0a0b8]/30"}`}
              >
                {persona ? "◆" : "●"}
              </span>
              <span className="truncate flex-1">{dmDisplayName(dm)}</span>
              {hasUnread && (
                <span className="ml-auto flex-shrink-0 bg-[#e2b714] text-[#0a0a23] text-[10px] font-bold rounded-full min-w-[18px] h-[18px] flex items-center justify-center px-1">
                  {unread > 99 ? "99+" : unread}
                </span>
              )}
            </button>
          );
        })}
      </div>

      {dms.length === 0 && !showPicker && (
        <p className="px-4 text-[#a0a0b8]/30 text-xs italic">no DMs yet</p>
      )}
    </div>
  );
}

export { dmDisplayName };
