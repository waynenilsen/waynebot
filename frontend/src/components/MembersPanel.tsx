import { useCallback, useEffect, useState } from "react";
import * as api from "../api";
import type { ChannelMember, Persona, User } from "../types";
import { useErrors } from "../store/ErrorContext";

interface MembersPanelProps {
  channelId: number;
  onClose: () => void;
}

export default function MembersPanel({
  channelId,
  onClose,
}: MembersPanelProps) {
  const [members, setMembers] = useState<ChannelMember[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [personas, setPersonas] = useState<Persona[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [showAdd, setShowAdd] = useState(false);
  const { pushError } = useErrors();

  const refresh = useCallback(async () => {
    try {
      const m = await api.getChannelMembers(channelId);
      setMembers(m);
    } catch (err) {
      pushError(
        `Failed to load members: ${err instanceof Error ? err.message : "unknown"}`,
      );
    }
  }, [channelId, pushError]);

  useEffect(() => {
    setLoading(true);
    Promise.all([
      api.getChannelMembers(channelId),
      api.getUsers(),
      api.getPersonas(),
    ])
      .then(([m, u, p]) => {
        setMembers(m);
        setUsers(u);
        setPersonas(p);
      })
      .catch((err) => {
        pushError(
          `Failed to load members: ${err instanceof Error ? err.message : "unknown"}`,
        );
      })
      .finally(() => setLoading(false));
  }, [channelId, pushError]);

  const addMember = useCallback(
    async (type: "user" | "persona", id: number) => {
      try {
        await api.addChannelMember(
          channelId,
          type === "user" ? { user_id: id } : { persona_id: id },
        );
        await refresh();
        setSearch("");
        setShowAdd(false);
      } catch (err) {
        pushError(
          `Failed to add member: ${err instanceof Error ? err.message : "unknown"}`,
        );
      }
    },
    [channelId, pushError, refresh],
  );

  const removeMember = useCallback(
    async (type: "user" | "persona", id: number) => {
      try {
        await api.removeChannelMember(
          channelId,
          type === "user" ? { user_id: id } : { persona_id: id },
        );
        await refresh();
      } catch (err) {
        pushError(
          `Failed to remove member: ${err instanceof Error ? err.message : "unknown"}`,
        );
      }
    },
    [channelId, pushError, refresh],
  );

  // Build list of addable users/personas (not already members)
  const memberUserIds = new Set(
    members.filter((m) => m.type === "user").map((m) => m.id),
  );
  const memberPersonaIds = new Set(
    members.filter((m) => m.type === "persona").map((m) => m.id),
  );

  const addable: { type: "user" | "persona"; id: number; name: string }[] = [];
  for (const u of users) {
    if (!memberUserIds.has(u.id)) {
      addable.push({ type: "user", id: u.id, name: u.username });
    }
  }
  for (const p of personas) {
    if (!memberPersonaIds.has(p.id)) {
      addable.push({ type: "persona", id: p.id, name: p.name });
    }
  }

  const lowerSearch = search.toLowerCase();
  const filtered = addable.filter((a) =>
    a.name.toLowerCase().includes(lowerSearch),
  );

  return (
    <div className="flex flex-col h-full bg-[#1a1a2e] border-l border-[#e2b714]/10 w-72">
      {/* Header */}
      <div className="shrink-0 px-4 py-3 border-b border-[#e2b714]/10 flex items-center justify-between">
        <h3 className="text-white text-sm font-bold font-mono">members</h3>
        <button
          onClick={onClose}
          className="text-[#a0a0b8]/50 hover:text-white text-xs font-mono transition-colors"
        >
          close
        </button>
      </div>

      {/* Member list */}
      <div className="flex-1 overflow-y-auto px-3 py-2">
        {loading ? (
          <p className="text-[#a0a0b8]/40 text-xs font-mono animate-pulse px-1 py-2">
            loading...
          </p>
        ) : members.length === 0 ? (
          <p className="text-[#a0a0b8]/40 text-xs font-mono px-1 py-2">
            no members
          </p>
        ) : (
          <ul className="space-y-1">
            {members.map((m) => (
              <li
                key={`${m.type}-${m.id}`}
                className="flex items-center justify-between group px-2 py-1.5 rounded hover:bg-[#0f3460]/30 transition-colors"
              >
                <div className="flex items-center gap-2 min-w-0">
                  <span
                    className={`shrink-0 w-1.5 h-1.5 rounded-full ${
                      m.type === "persona"
                        ? "bg-[#e2b714]/60"
                        : "bg-[#a0a0b8]/40"
                    }`}
                  />
                  <span className="text-white text-xs font-mono truncate">
                    {m.name}
                  </span>
                  {m.role === "owner" && (
                    <span className="text-[#e2b714]/40 text-[10px] font-mono shrink-0">
                      owner
                    </span>
                  )}
                  {m.type === "persona" && (
                    <span className="text-[#a0a0b8]/30 text-[10px] font-mono shrink-0">
                      agent
                    </span>
                  )}
                </div>
                {m.role !== "owner" && (
                  <button
                    onClick={() => removeMember(m.type, m.id)}
                    className="text-[#a0a0b8]/20 hover:text-red-400 text-xs font-mono opacity-0 group-hover:opacity-100 transition-all shrink-0 ml-2"
                    title="remove"
                  >
                    x
                  </button>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Add member section */}
      <div className="shrink-0 border-t border-[#e2b714]/10 px-3 py-2">
        {!showAdd ? (
          <button
            onClick={() => setShowAdd(true)}
            className="w-full text-left text-[#e2b714]/60 hover:text-[#e2b714] text-xs font-mono py-1.5 px-2 rounded hover:bg-[#0f3460]/30 transition-colors"
          >
            + add member
          </button>
        ) : (
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <input
                type="text"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="search users & agents..."
                autoFocus
                className="flex-1 bg-[#0f3460]/30 text-white text-xs font-mono px-2 py-1.5 rounded border border-[#e2b714]/10 focus:border-[#e2b714]/30 outline-none placeholder:text-[#a0a0b8]/30"
              />
              <button
                onClick={() => {
                  setShowAdd(false);
                  setSearch("");
                }}
                className="text-[#a0a0b8]/40 hover:text-white text-xs font-mono transition-colors"
              >
                x
              </button>
            </div>
            <ul className="max-h-32 overflow-y-auto space-y-0.5">
              {filtered.length === 0 ? (
                <li className="text-[#a0a0b8]/30 text-xs font-mono px-2 py-1">
                  {addable.length === 0
                    ? "everyone is already a member"
                    : "no matches"}
                </li>
              ) : (
                filtered.map((a) => (
                  <li key={`${a.type}-${a.id}`}>
                    <button
                      onClick={() => addMember(a.type, a.id)}
                      className="w-full text-left flex items-center gap-2 px-2 py-1.5 rounded hover:bg-[#0f3460]/30 transition-colors"
                    >
                      <span
                        className={`shrink-0 w-1.5 h-1.5 rounded-full ${
                          a.type === "persona"
                            ? "bg-[#e2b714]/60"
                            : "bg-[#a0a0b8]/40"
                        }`}
                      />
                      <span className="text-white text-xs font-mono truncate">
                        {a.name}
                      </span>
                      {a.type === "persona" && (
                        <span className="text-[#a0a0b8]/30 text-[10px] font-mono">
                          agent
                        </span>
                      )}
                    </button>
                  </li>
                ))
              )}
            </ul>
          </div>
        )}
      </div>
    </div>
  );
}
