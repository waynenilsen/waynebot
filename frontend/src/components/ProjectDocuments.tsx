import { useState, useEffect, useCallback } from "react";
import {
  getProjectDocuments,
  getProjectDocument,
  updateProjectDocument,
  appendToDocument,
  deleteDocument,
} from "../api";
import type { ProjectDocumentList } from "../types";

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
  const [docs, setDocs] = useState<ProjectDocumentList[]>([]);
  const [selectedFile, setSelectedFile] = useState<string | null>(null);
  const [content, setContent] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [dirty, setDirty] = useState(false);
  const [newFileName, setNewFileName] = useState("");
  const [appendInput, setAppendInput] = useState("");
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

  const loadFile = useCallback(
    async (type: DocType, filename: string) => {
      try {
        const doc = await getProjectDocument(projectId, type, filename);
        setContent(doc.content || "");
        setDirty(false);
      } catch {
        setContent("");
        setDirty(false);
      }
    },
    [projectId],
  );

  useEffect(() => {
    loadDocs();
  }, [loadDocs]);

  useEffect(() => {
    setSelectedFile(null);
    setContent("");
    setDirty(false);
  }, [tab]);

  useEffect(() => {
    if (selectedFile) {
      loadFile(tab, selectedFile);
    }
  }, [selectedFile, tab, loadFile]);

  const currentFiles = docs.find((d) => d.type === tab)?.files ?? [];

  async function handleSave() {
    if (!selectedFile) return;
    setSaving(true);
    setError(null);
    try {
      await updateProjectDocument(projectId, tab, selectedFile, content);
      setDirty(false);
      await loadDocs();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to save");
    } finally {
      setSaving(false);
    }
  }

  async function handleCreateFile() {
    const name = newFileName.trim();
    if (!name) return;
    const filename = name.endsWith(".md") ? name : name + ".md";
    setSaving(true);
    setError(null);
    try {
      await updateProjectDocument(projectId, tab, filename, "");
      setNewFileName("");
      await loadDocs();
      setSelectedFile(filename);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to create file");
    } finally {
      setSaving(false);
    }
  }

  async function handleDeleteFile(filename: string) {
    setSaving(true);
    setError(null);
    try {
      await deleteDocument(projectId, tab, filename);
      if (selectedFile === filename) {
        setSelectedFile(null);
        setContent("");
      }
      await loadDocs();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to delete file");
    } finally {
      setSaving(false);
    }
  }

  async function handleAppend() {
    if (!selectedFile) return;
    const text = appendInput.trim();
    if (!text) return;
    setSaving(true);
    setError(null);
    try {
      await appendToDocument(projectId, tab, selectedFile, text);
      setAppendInput("");
      await loadFile(tab, selectedFile);
      await loadDocs();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to append");
    } finally {
      setSaving(false);
    }
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
        {DOC_TYPES.map((dt) => {
          const files = docs.find((d) => d.type === dt.key)?.files ?? [];
          return (
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
              {files.length > 0 && (
                <span className="text-[10px] text-[#a0a0b8]/40">
                  ({files.length})
                </span>
              )}
            </button>
          );
        })}
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
        ) : selectedFile ? (
          /* File editor view */
          <>
            <div className="flex items-center gap-2 shrink-0">
              <button
                onClick={() => {
                  setSelectedFile(null);
                  setContent("");
                  setDirty(false);
                }}
                className="text-[#a0a0b8]/40 hover:text-white text-xs px-2 py-1 rounded hover:bg-white/5 transition-colors cursor-pointer"
              >
                &larr; back
              </button>
              <span className="text-[#e2b714] text-xs font-mono">
                {tab}/{selectedFile}
              </span>
            </div>

            {tab === "decisions" ? (
              /* Decisions: show content read-only + append input */
              <>
                <div className="flex-1 overflow-y-auto">
                  <pre className="text-[#c8c8e0] text-sm font-mono whitespace-pre-wrap">
                    {content || "empty"}
                  </pre>
                </div>
                <div className="flex gap-2 shrink-0">
                  <input
                    value={appendInput}
                    onChange={(e) => setAppendInput(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter" && !e.shiftKey) {
                        e.preventDefault();
                        handleAppend();
                      }
                    }}
                    placeholder="Add an entry..."
                    className="flex-1 bg-[#0f3460]/50 border border-[#e2b714]/10 rounded px-3 py-2.5 text-[#c8c8e0] text-sm font-mono focus:outline-none focus:border-[#e2b714]/30 placeholder:text-[#a0a0b8]/20"
                  />
                  <button
                    onClick={handleAppend}
                    disabled={saving || !appendInput.trim()}
                    className={`text-sm font-semibold py-2 px-4 rounded transition-colors cursor-pointer shrink-0 ${
                      appendInput.trim()
                        ? "bg-[#e2b714] hover:bg-[#c9a212] text-[#1a1a2e]"
                        : "bg-[#e2b714]/20 text-[#e2b714]/40 cursor-not-allowed"
                    }`}
                  >
                    {saving ? "Adding..." : "Append"}
                  </button>
                </div>
              </>
            ) : (
              /* ERD/PRD: editable textarea */
              <>
                <textarea
                  value={content}
                  onChange={(e) => {
                    setContent(e.target.value);
                    setDirty(true);
                  }}
                  placeholder="Write markdown here..."
                  className="flex-1 w-full bg-[#0f3460]/50 border border-[#e2b714]/10 rounded px-3 py-2.5 text-[#c8c8e0] text-sm font-mono resize-none focus:outline-none focus:border-[#e2b714]/30 placeholder:text-[#a0a0b8]/20"
                />
                <button
                  onClick={handleSave}
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
            )}
          </>
        ) : (
          /* File list view */
          <>
            {currentFiles.length === 0 ? (
              <div className="text-[#a0a0b8]/30 text-xs font-mono text-center py-8">
                no documents yet
              </div>
            ) : (
              <div className="flex-1 overflow-y-auto space-y-1">
                {currentFiles.map((f) => (
                  <div
                    key={f}
                    className="flex items-center justify-between bg-[#0f3460]/30 border border-[#e2b714]/5 rounded px-3 py-2 group"
                  >
                    <button
                      onClick={() => setSelectedFile(f)}
                      className="text-[#c8c8e0] text-sm font-mono hover:text-[#e2b714] transition-colors cursor-pointer text-left flex-1 truncate"
                    >
                      {f}
                    </button>
                    <button
                      onClick={() => handleDeleteFile(f)}
                      className="text-[#a0a0b8]/20 hover:text-red-400 text-xs px-2 py-0.5 rounded hover:bg-red-500/5 transition-colors cursor-pointer opacity-0 group-hover:opacity-100"
                    >
                      delete
                    </button>
                  </div>
                ))}
              </div>
            )}

            {/* Create new file */}
            <div className="flex gap-2 shrink-0">
              <input
                value={newFileName}
                onChange={(e) => setNewFileName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault();
                    handleCreateFile();
                  }
                }}
                placeholder="new-document.md"
                className="flex-1 bg-[#0f3460]/50 border border-[#e2b714]/10 rounded px-3 py-2.5 text-[#c8c8e0] text-sm font-mono focus:outline-none focus:border-[#e2b714]/30 placeholder:text-[#a0a0b8]/20"
              />
              <button
                onClick={handleCreateFile}
                disabled={saving || !newFileName.trim()}
                className={`text-sm font-semibold py-2 px-4 rounded transition-colors cursor-pointer shrink-0 ${
                  newFileName.trim()
                    ? "bg-[#e2b714] hover:bg-[#c9a212] text-[#1a1a2e]"
                    : "bg-[#e2b714]/20 text-[#e2b714]/40 cursor-not-allowed"
                }`}
              >
                {saving ? "Creating..." : "Create"}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
