import { useState, useEffect, useCallback } from "react";
import {
  getProjectDocuments,
  getProjectDocument,
  updateProjectDocument,
  appendDecision,
} from "../api";
import type { ProjectDocument } from "../types";

const DOC_TYPES = [
  { key: "erd", label: "ERD" },
  { key: "prd", label: "PRD" },
  { key: "decisions", label: "Decisions" },
] as const;

type DocType = (typeof DOC_TYPES)[number]["key"];

interface Props {
  projectId: number;
  onClose: () => void;
}

export default function ProjectDocuments({ projectId, onClose }: Props) {
  const [tab, setTab] = useState<DocType>("erd");
  const [docs, setDocs] = useState<ProjectDocument[]>([]);
  const [content, setContent] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [dirty, setDirty] = useState(false);
  const [decisionInput, setDecisionInput] = useState("");
  const [error, setError] = useState<string | null>(null);

  const loadDocs = useCallback(async () => {
    setLoading(true);
    try {
      const list = await getProjectDocuments(projectId);
      setDocs(list);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load documents");
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  const loadDoc = useCallback(
    async (type: DocType) => {
      try {
        const doc = await getProjectDocument(projectId, type);
        setContent(doc.content || "");
        setDirty(false);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to load document");
      }
    },
    [projectId],
  );

  useEffect(() => {
    loadDocs();
  }, [loadDocs]);

  useEffect(() => {
    loadDoc(tab);
  }, [tab, loadDoc]);

  async function handleSave() {
    setSaving(true);
    setError(null);
    try {
      await updateProjectDocument(projectId, tab, content);
      setDirty(false);
      await loadDocs();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to save");
    } finally {
      setSaving(false);
    }
  }

  async function handleAddDecision() {
    const text = decisionInput.trim();
    if (!text) return;
    setSaving(true);
    setError(null);
    try {
      await appendDecision(projectId, text);
      setDecisionInput("");
      await loadDoc("decisions");
      await loadDocs();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to add decision");
    } finally {
      setSaving(false);
    }
  }

  function docStatus(type: string): boolean {
    const d = docs.find((doc) => doc.type === type);
    return d?.exists ?? false;
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-[#e2b714]/10 shrink-0">
        <h2 className="text-white text-sm font-bold font-mono">Documents</h2>
        <button
          onClick={onClose}
          className="text-[#a0a0b8]/40 hover:text-white text-xs px-2 py-1 rounded hover:bg-white/5 transition-colors cursor-pointer"
        >
          close
        </button>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-[#e2b714]/10 shrink-0">
        {DOC_TYPES.map((dt) => (
          <button
            key={dt.key}
            onClick={() => setTab(dt.key)}
            className={`flex-1 text-xs font-mono py-2.5 px-3 transition-colors cursor-pointer flex items-center justify-center gap-1.5 ${
              tab === dt.key
                ? "text-[#e2b714] border-b-2 border-[#e2b714] bg-[#e2b714]/5"
                : "text-[#a0a0b8]/50 hover:text-[#a0a0b8] hover:bg-white/[0.02]"
            }`}
          >
            {dt.label}
            <span
              className={`w-1.5 h-1.5 rounded-full ${docStatus(dt.key) ? "bg-[#e2b714]/60" : "bg-[#a0a0b8]/20"}`}
            />
          </button>
        ))}
      </div>

      {/* Error */}
      {error && (
        <div className="mx-4 mt-3 px-3 py-2 bg-red-500/10 border border-red-500/20 rounded text-red-400 text-xs font-mono">
          {error}
        </div>
      )}

      {/* Content */}
      <div className="flex-1 flex flex-col overflow-hidden p-4 gap-3">
        {loading ? (
          <div className="text-[#a0a0b8]/50 text-xs font-mono text-center py-8">
            loading...
          </div>
        ) : tab === "decisions" ? (
          <DecisionsView
            content={content}
            decisionInput={decisionInput}
            onInputChange={setDecisionInput}
            onAdd={handleAddDecision}
            saving={saving}
          />
        ) : (
          <MarkdownEditor
            content={content}
            onChange={(v) => {
              setContent(v);
              setDirty(true);
            }}
            onSave={handleSave}
            saving={saving}
            dirty={dirty}
          />
        )}
      </div>
    </div>
  );
}

function MarkdownEditor({
  content,
  onChange,
  onSave,
  saving,
  dirty,
}: {
  content: string;
  onChange: (v: string) => void;
  onSave: () => void;
  saving: boolean;
  dirty: boolean;
}) {
  return (
    <>
      <textarea
        value={content}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Write markdown here..."
        className="flex-1 w-full bg-[#0f3460]/50 border border-[#e2b714]/10 rounded px-3 py-2.5 text-[#c8c8e0] text-sm font-mono resize-none focus:outline-none focus:border-[#e2b714]/30 placeholder:text-[#a0a0b8]/20"
      />
      <button
        onClick={onSave}
        disabled={saving || !dirty}
        className={`self-end text-sm font-semibold py-2 px-4 rounded transition-colors cursor-pointer ${
          dirty
            ? "bg-[#e2b714] hover:bg-[#c9a212] text-[#1a1a2e]"
            : "bg-[#e2b714]/20 text-[#e2b714]/40 cursor-not-allowed"
        }`}
      >
        {saving ? "Saving..." : "Save"}
      </button>
    </>
  );
}

function DecisionsView({
  content,
  decisionInput,
  onInputChange,
  onAdd,
  saving,
}: {
  content: string;
  decisionInput: string;
  onInputChange: (v: string) => void;
  onAdd: () => void;
  saving: boolean;
}) {
  const decisions = content
    ? content
        .split("\n")
        .filter((l) => l.trim())
        .map((line) => {
          const match = line.match(/^\[(.+?)\]\s*(.*)/);
          if (match) return { timestamp: match[1], text: match[2] };
          return { timestamp: "", text: line };
        })
    : [];

  return (
    <>
      {/* Existing decisions */}
      <div className="flex-1 overflow-y-auto space-y-2">
        {decisions.length === 0 ? (
          <div className="text-[#a0a0b8]/30 text-xs font-mono text-center py-8">
            no decisions recorded yet
          </div>
        ) : (
          decisions.map((d, i) => (
            <div
              key={i}
              className="bg-[#0f3460]/30 border border-[#e2b714]/5 rounded px-3 py-2"
            >
              {d.timestamp && (
                <span className="text-[#a0a0b8]/30 text-[10px] font-mono block mb-0.5">
                  {d.timestamp}
                </span>
              )}
              <span className="text-[#c8c8e0] text-sm font-mono">
                {d.text}
              </span>
            </div>
          ))
        )}
      </div>

      {/* Add decision */}
      <div className="flex gap-2 shrink-0">
        <input
          value={decisionInput}
          onChange={(e) => onInputChange(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && !e.shiftKey) {
              e.preventDefault();
              onAdd();
            }
          }}
          placeholder="Add a decision..."
          className="flex-1 bg-[#0f3460]/50 border border-[#e2b714]/10 rounded px-3 py-2.5 text-[#c8c8e0] text-sm font-mono focus:outline-none focus:border-[#e2b714]/30 placeholder:text-[#a0a0b8]/20"
        />
        <button
          onClick={onAdd}
          disabled={saving || !decisionInput.trim()}
          className={`text-sm font-semibold py-2 px-4 rounded transition-colors cursor-pointer shrink-0 ${
            decisionInput.trim()
              ? "bg-[#e2b714] hover:bg-[#c9a212] text-[#1a1a2e]"
              : "bg-[#e2b714]/20 text-[#e2b714]/40 cursor-not-allowed"
          }`}
        >
          {saving ? "Adding..." : "Add Decision"}
        </button>
      </div>
    </>
  );
}
