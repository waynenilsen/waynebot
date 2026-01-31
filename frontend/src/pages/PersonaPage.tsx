import { useState } from "react";
import { usePersonas } from "../hooks/usePersonas";
import PersonaForm from "../components/PersonaForm";
import type { Persona } from "../types";

type PersonaData = Omit<Persona, "id" | "created_at">;

export default function PersonaPage() {
  const { personas, loading, createPersona, updatePersona, deletePersona } =
    usePersonas();
  const [editing, setEditing] = useState<Persona | null>(null);
  const [creating, setCreating] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState<number | null>(null);

  const showForm = creating || editing !== null;

  async function handleCreate(data: PersonaData) {
    await createPersona(data);
    setCreating(false);
  }

  async function handleUpdate(data: PersonaData) {
    if (!editing) return;
    await updatePersona(editing.id, data);
    setEditing(null);
  }

  async function handleDelete(id: number) {
    await deletePersona(id);
    setConfirmDelete(null);
  }

  if (showForm) {
    return (
      <div className="flex-1 overflow-y-auto p-6">
        <div className="max-w-2xl mx-auto">
          <PersonaForm
            initial={editing ?? undefined}
            onSubmit={editing ? handleUpdate : handleCreate}
            onCancel={() => {
              setEditing(null);
              setCreating(false);
            }}
          />
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* Header */}
      <div className="px-6 py-4 border-b border-[#e2b714]/10 flex items-center justify-between shrink-0">
        <div>
          <h1 className="text-white text-lg font-bold font-mono flex items-center gap-2">
            <span className="text-[#e2b714]/40 text-sm">&#9670;</span>
            Personas
          </h1>
          <p className="text-[#a0a0b8]/50 text-xs font-mono mt-0.5">
            {personas.length} persona{personas.length !== 1 && "s"} configured
          </p>
        </div>
        <button
          onClick={() => setCreating(true)}
          className="bg-[#e2b714] hover:bg-[#c9a212] text-[#1a1a2e] font-semibold text-sm py-2 px-4 rounded transition-colors cursor-pointer"
        >
          + New Persona
        </button>
      </div>

      {/* List */}
      <div className="flex-1 overflow-y-auto p-6">
        {loading && personas.length === 0 ? (
          <div className="text-[#a0a0b8]/50 text-sm font-mono text-center py-12">
            loading...
          </div>
        ) : personas.length === 0 ? (
          <div className="text-[#a0a0b8]/50 text-sm font-mono text-center py-12">
            no personas yet — create one to get started
          </div>
        ) : (
          <div className="space-y-3 max-w-3xl mx-auto">
            {personas.map((p) => (
              <div
                key={p.id}
                className="bg-[#16213e] border border-[#e2b714]/10 rounded-lg p-4 group"
              >
                <div className="flex items-start justify-between gap-4">
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <div className="w-6 h-6 rotate-45 bg-[#e2b714]/10 border border-[#e2b714]/25 flex items-center justify-center shrink-0">
                        <span className="-rotate-45 text-[#e2b714] text-[10px] font-bold uppercase">
                          {p.name.charAt(0)}
                        </span>
                      </div>
                      <h3 className="text-white font-bold font-mono text-sm truncate">
                        {p.name}
                      </h3>
                      <span className="text-[#a0a0b8]/30 text-xs font-mono shrink-0">
                        {p.model}
                      </span>
                    </div>
                    <p className="text-[#a0a0b8]/50 text-xs font-mono mt-2 line-clamp-2">
                      {p.system_prompt}
                    </p>
                    <div className="flex items-center gap-3 mt-2 text-[10px] text-[#a0a0b8]/30 font-mono uppercase tracking-wider">
                      <span>temp {p.temperature}</span>
                      <span className="text-[#e2b714]/15">|</span>
                      <span>{p.max_tokens.toLocaleString()} tokens</span>
                      <span className="text-[#e2b714]/15">|</span>
                      <span>{p.cooldown_secs}s cooldown</span>
                      {p.tools_enabled.length > 0 && (
                        <>
                          <span className="text-[#e2b714]/15">|</span>
                          <span>{p.tools_enabled.length} tools</span>
                        </>
                      )}
                    </div>
                  </div>

                  <div className="flex items-center gap-1 shrink-0">
                    <button
                      onClick={() => setEditing(p)}
                      className="text-[#a0a0b8]/40 hover:text-[#e2b714] text-xs px-2 py-1 rounded hover:bg-[#e2b714]/5 transition-colors cursor-pointer"
                    >
                      edit
                    </button>
                    {confirmDelete === p.id ? (
                      <span className="flex items-center gap-1">
                        <button
                          onClick={() => handleDelete(p.id)}
                          className="text-red-400 text-xs px-2 py-1 rounded hover:bg-red-500/10 transition-colors cursor-pointer"
                        >
                          confirm
                        </button>
                        <button
                          onClick={() => setConfirmDelete(null)}
                          className="text-[#a0a0b8]/40 text-xs px-2 py-1 rounded hover:bg-white/5 transition-colors cursor-pointer"
                        >
                          cancel
                        </button>
                      </span>
                    ) : (
                      <button
                        onClick={() => setConfirmDelete(p.id)}
                        className="text-[#a0a0b8]/40 hover:text-red-400 text-xs px-2 py-1 rounded hover:bg-red-500/5 transition-colors cursor-pointer"
                      >
                        delete
                      </button>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
