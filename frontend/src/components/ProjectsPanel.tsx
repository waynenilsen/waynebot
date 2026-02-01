import { useCallback, useEffect, useState } from "react";
import * as api from "../api";
import type { Project } from "../types";
import { useErrors } from "../store/ErrorContext";
import { getErrorMessage } from "../utils/errors";

interface ProjectsPanelProps {
  channelId: number;
  onClose: () => void;
}

export default function ProjectsPanel({
  channelId,
  onClose,
}: ProjectsPanelProps) {
  const [channelProjects, setChannelProjects] = useState<Project[]>([]);
  const [allProjects, setAllProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [showAdd, setShowAdd] = useState(false);
  const { pushError } = useErrors();

  const refresh = useCallback(async () => {
    try {
      const cp = await api.getChannelProjects(channelId);
      setChannelProjects(cp);
    } catch (err) {
      pushError(
        `Failed to load projects: ${getErrorMessage(err)}`,
      );
    }
  }, [channelId, pushError]);

  useEffect(() => {
    setLoading(true);
    Promise.all([api.getChannelProjects(channelId), api.getProjects()])
      .then(([cp, ap]) => {
        setChannelProjects(cp);
        setAllProjects(ap);
      })
      .catch((err) => {
        pushError(
          `Failed to load projects: ${getErrorMessage(err)}`,
        );
      })
      .finally(() => setLoading(false));
  }, [channelId, pushError]);

  const addProject = useCallback(
    async (projectId: number) => {
      try {
        await api.addChannelProject(channelId, projectId);
        await refresh();
        setSearch("");
        setShowAdd(false);
      } catch (err) {
        pushError(
          `Failed to add project: ${getErrorMessage(err)}`,
        );
      }
    },
    [channelId, pushError, refresh],
  );

  const removeProject = useCallback(
    async (projectId: number) => {
      try {
        await api.removeChannelProject(channelId, projectId);
        await refresh();
      } catch (err) {
        pushError(
          `Failed to remove project: ${getErrorMessage(err)}`,
        );
      }
    },
    [channelId, pushError, refresh],
  );

  const associatedIds = new Set(channelProjects.map((p) => p.id));
  const addable = allProjects.filter((p) => !associatedIds.has(p.id));
  const lowerSearch = search.toLowerCase();
  const filtered = addable.filter((p) =>
    p.name.toLowerCase().includes(lowerSearch),
  );

  return (
    <div className="flex flex-col h-full bg-[#1a1a2e] border-l border-[#e2b714]/10 w-72">
      {/* Header */}
      <div className="shrink-0 px-4 py-3 border-b border-[#e2b714]/10 flex items-center justify-between">
        <h3 className="text-white text-sm font-bold font-mono">projects</h3>
        <button
          onClick={onClose}
          className="text-[#a0a0b8]/50 hover:text-white text-xs font-mono transition-colors"
        >
          close
        </button>
      </div>

      {/* Project list */}
      <div className="flex-1 overflow-y-auto px-3 py-2">
        {loading ? (
          <p className="text-[#a0a0b8]/40 text-xs font-mono animate-pulse px-1 py-2">
            loading...
          </p>
        ) : channelProjects.length === 0 ? (
          <p className="text-[#a0a0b8]/40 text-xs font-mono px-1 py-2">
            no projects associated
          </p>
        ) : (
          <ul className="space-y-1">
            {channelProjects.map((p) => (
              <li
                key={p.id}
                className="flex items-center justify-between group px-2 py-1.5 rounded hover:bg-[#0f3460]/30 transition-colors"
              >
                <div className="flex items-center gap-2 min-w-0">
                  <span className="shrink-0 w-1.5 h-1.5 rounded-sm bg-[#e2b714]/60" />
                  <div className="min-w-0">
                    <span className="text-white text-xs font-mono truncate block">
                      {p.name}
                    </span>
                    <span className="text-[#a0a0b8]/30 text-[10px] font-mono truncate block">
                      {p.path}
                    </span>
                  </div>
                </div>
                <button
                  onClick={() => removeProject(p.id)}
                  className="text-[#a0a0b8]/20 hover:text-red-400 text-xs font-mono opacity-0 group-hover:opacity-100 transition-all shrink-0 ml-2"
                  title="remove"
                >
                  x
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Add project section */}
      <div className="shrink-0 border-t border-[#e2b714]/10 px-3 py-2">
        {!showAdd ? (
          <button
            onClick={() => setShowAdd(true)}
            className="w-full text-left text-[#e2b714]/60 hover:text-[#e2b714] text-xs font-mono py-1.5 px-2 rounded hover:bg-[#0f3460]/30 transition-colors"
          >
            + add project
          </button>
        ) : (
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <input
                type="text"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="search projects..."
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
                    ? "all projects already associated"
                    : "no matches"}
                </li>
              ) : (
                filtered.map((p) => (
                  <li key={p.id}>
                    <button
                      onClick={() => addProject(p.id)}
                      className="w-full text-left flex items-center gap-2 px-2 py-1.5 rounded hover:bg-[#0f3460]/30 transition-colors"
                    >
                      <span className="shrink-0 w-1.5 h-1.5 rounded-sm bg-[#e2b714]/60" />
                      <span className="text-white text-xs font-mono truncate">
                        {p.name}
                      </span>
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
