import { useEffect, useState, useCallback } from "react";
import { useAgentActivity } from "../hooks/useAgentActivity";
import type { AgentStatus, LLMCall, ToolExecution } from "../types";

type ActivityEntry =
  | { kind: "llm"; data: LLMCall }
  | { kind: "tool"; data: ToolExecution };

function mergeActivity(
  calls: LLMCall[],
  execs: ToolExecution[],
): ActivityEntry[] {
  const entries: ActivityEntry[] = [
    ...calls.map((c) => ({ kind: "llm" as const, data: c })),
    ...execs.map((e) => ({ kind: "tool" as const, data: e })),
  ];
  entries.sort(
    (a, b) =>
      new Date(b.data.created_at).getTime() -
      new Date(a.data.created_at).getTime(),
  );
  return entries;
}

function formatTime(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function formatDate(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleDateString([], { month: "short", day: "numeric" });
}

function LLMCallEntry({ call }: { call: LLMCall }) {
  const [expanded, setExpanded] = useState(false);
  const tokens = call.prompt_tokens + call.completion_tokens;

  return (
    <div className="border border-[#e2b714]/10 rounded-lg bg-[#16213e] overflow-hidden">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full text-left px-4 py-3 flex items-center gap-3 hover:bg-[#e2b714]/5 transition-colors cursor-pointer"
      >
        <span className="text-blue-400 text-xs font-mono font-bold bg-blue-400/10 px-2 py-0.5 rounded shrink-0">
          LLM
        </span>
        <span className="text-white/70 text-sm font-mono truncate flex-1">
          {call.model}
        </span>
        <span className="text-[#a0a0b8]/40 text-xs font-mono shrink-0">
          {tokens} tok
        </span>
        <span className="text-[#a0a0b8]/30 text-xs font-mono shrink-0">
          {formatDate(call.created_at)} {formatTime(call.created_at)}
        </span>
        <span className="text-[#a0a0b8]/30 text-xs">
          {expanded ? "\u25BC" : "\u25B6"}
        </span>
      </button>
      {expanded && (
        <div className="border-t border-[#e2b714]/10 px-4 py-3 space-y-3">
          <div>
            <p className="text-[#a0a0b8]/50 text-[10px] uppercase tracking-widest mb-1">
              Tokens
            </p>
            <p className="text-[#a0a0b8]/70 text-xs font-mono">
              prompt: {call.prompt_tokens} | completion:{" "}
              {call.completion_tokens}
            </p>
          </div>
          <div>
            <p className="text-[#a0a0b8]/50 text-[10px] uppercase tracking-widest mb-1">
              Response
            </p>
            <pre className="text-[#a0a0b8]/60 text-xs font-mono bg-[#0f3460]/30 p-3 rounded overflow-x-auto max-h-64 overflow-y-auto whitespace-pre-wrap break-words">
              {formatJSON(call.response_json)}
            </pre>
          </div>
          <div>
            <p className="text-[#a0a0b8]/50 text-[10px] uppercase tracking-widest mb-1">
              Messages
            </p>
            <pre className="text-[#a0a0b8]/60 text-xs font-mono bg-[#0f3460]/30 p-3 rounded overflow-x-auto max-h-64 overflow-y-auto whitespace-pre-wrap break-words">
              {formatJSON(call.messages_json)}
            </pre>
          </div>
        </div>
      )}
    </div>
  );
}

function ToolExecEntry({ exec }: { exec: ToolExecution }) {
  const [expanded, setExpanded] = useState(false);
  const hasError = exec.error_text !== "";

  return (
    <div
      className={`border rounded-lg bg-[#16213e] overflow-hidden ${
        hasError ? "border-red-500/20" : "border-[#e2b714]/10"
      }`}
    >
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full text-left px-4 py-3 flex items-center gap-3 hover:bg-[#e2b714]/5 transition-colors cursor-pointer"
      >
        <span
          className={`text-xs font-mono font-bold px-2 py-0.5 rounded shrink-0 ${
            hasError
              ? "text-red-400 bg-red-400/10"
              : "text-emerald-400 bg-emerald-400/10"
          }`}
        >
          TOOL
        </span>
        <span className="text-white/70 text-sm font-mono truncate flex-1">
          {exec.tool_name}
        </span>
        <span className="text-[#a0a0b8]/40 text-xs font-mono shrink-0">
          {exec.duration_ms}ms
        </span>
        <span className="text-[#a0a0b8]/30 text-xs font-mono shrink-0">
          {formatDate(exec.created_at)} {formatTime(exec.created_at)}
        </span>
        <span className="text-[#a0a0b8]/30 text-xs">
          {expanded ? "\u25BC" : "\u25B6"}
        </span>
      </button>
      {expanded && (
        <div className="border-t border-[#e2b714]/10 px-4 py-3 space-y-3">
          <div>
            <p className="text-[#a0a0b8]/50 text-[10px] uppercase tracking-widest mb-1">
              Arguments
            </p>
            <pre className="text-[#a0a0b8]/60 text-xs font-mono bg-[#0f3460]/30 p-3 rounded overflow-x-auto max-h-40 overflow-y-auto whitespace-pre-wrap break-words">
              {formatJSON(exec.args_json)}
            </pre>
          </div>
          {exec.output_text && (
            <div>
              <p className="text-[#a0a0b8]/50 text-[10px] uppercase tracking-widest mb-1">
                Output
              </p>
              <pre className="text-[#a0a0b8]/60 text-xs font-mono bg-[#0f3460]/30 p-3 rounded overflow-x-auto max-h-40 overflow-y-auto whitespace-pre-wrap break-words">
                {exec.output_text}
              </pre>
            </div>
          )}
          {hasError && (
            <div>
              <p className="text-red-400/70 text-[10px] uppercase tracking-widest mb-1">
                Error
              </p>
              <pre className="text-red-400/70 text-xs font-mono bg-red-500/5 p-3 rounded overflow-x-auto whitespace-pre-wrap break-words">
                {exec.error_text}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function formatJSON(raw: string): string {
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}

interface Props {
  agent: AgentStatus;
  onBack: () => void;
}

export default function AgentActivityPage({ agent, onBack }: Props) {
  const {
    llmCalls,
    toolExecs,
    stats,
    loading,
    hasMoreCalls,
    hasMoreExecs,
    loadActivity,
    loadMoreCalls,
    loadMoreExecs,
    refreshStats,
    prependLLMCall,
    prependToolExec,
  } = useAgentActivity();

  const [tab, setTab] = useState<"all" | "llm" | "tools">("all");

  useEffect(() => {
    loadActivity(agent.persona_id);
  }, [agent.persona_id, loadActivity]);

  // Poll stats every 10 seconds when agent is active
  useEffect(() => {
    const active = ["idle", "thinking", "tool_call"].includes(agent.status);
    if (!active) return;

    const interval = setInterval(() => {
      refreshStats(agent.persona_id);
    }, 10000);
    return () => clearInterval(interval);
  }, [agent.persona_id, agent.status, refreshStats]);

  // Live WebSocket updates
  useEffect(() => {
    const handleLLMCall = (e: Event) => {
      const call = (e as CustomEvent).detail as LLMCall;
      if (call.persona_id === agent.persona_id) {
        prependLLMCall(call);
      }
    };
    const handleToolExec = (e: Event) => {
      const exec = (e as CustomEvent).detail as ToolExecution;
      if (exec.persona_id === agent.persona_id) {
        prependToolExec(exec);
      }
    };

    window.addEventListener("agent_llm_call", handleLLMCall);
    window.addEventListener("agent_tool_execution", handleToolExec);
    return () => {
      window.removeEventListener("agent_llm_call", handleLLMCall);
      window.removeEventListener("agent_tool_execution", handleToolExec);
    };
  }, [agent.persona_id, prependLLMCall, prependToolExec]);

  const handleLoadMore = useCallback(() => {
    if (tab === "llm" || tab === "all") {
      loadMoreCalls(agent.persona_id);
    }
    if (tab === "tools" || tab === "all") {
      loadMoreExecs(agent.persona_id);
    }
  }, [tab, agent.persona_id, loadMoreCalls, loadMoreExecs]);

  const activity = mergeActivity(
    tab === "tools" ? [] : llmCalls,
    tab === "llm" ? [] : toolExecs,
  );

  const canLoadMore =
    (tab !== "tools" && hasMoreCalls) || (tab !== "llm" && hasMoreExecs);

  const isActive = ["idle", "thinking", "tool_call"].includes(agent.status);

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* Header */}
      <div className="px-6 py-4 border-b border-[#e2b714]/10 shrink-0">
        <div className="flex items-center gap-3 mb-2">
          <button
            onClick={onBack}
            className="text-[#a0a0b8]/60 hover:text-[#a0a0b8] text-sm font-mono cursor-pointer"
          >
            &larr; agents
          </button>
          <div
            className={`w-2.5 h-2.5 rounded-full shrink-0 ${
              isActive
                ? "bg-emerald-400 shadow-[0_0_6px_rgba(52,211,153,0.4)]"
                : "bg-[#a0a0b8]/30"
            }`}
          />
          <h1 className="text-white text-lg font-bold font-mono">
            {agent.persona_name}
          </h1>
          <span
            className={`text-[10px] uppercase tracking-widest font-mono ${
              isActive ? "text-emerald-400/60" : "text-[#a0a0b8]/30"
            }`}
          >
            {agent.status}
          </span>
        </div>

        {/* Stats bar */}
        {stats && (
          <div className="flex items-center gap-6 text-xs font-mono">
            <div className="flex items-center gap-1.5">
              <span className="text-[#a0a0b8]/40">calls/hr:</span>
              <span className="text-white/70">
                {stats.total_calls_last_hour}
              </span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="text-[#a0a0b8]/40">tokens/hr:</span>
              <span className="text-white/70">
                {stats.total_tokens_last_hour.toLocaleString()}
              </span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="text-[#a0a0b8]/40">errors/hr:</span>
              <span
                className={
                  stats.error_count_last_hour > 0
                    ? "text-red-400"
                    : "text-white/70"
                }
              >
                {stats.error_count_last_hour}
              </span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="text-[#a0a0b8]/40">avg tool:</span>
              <span className="text-white/70">
                {Math.round(stats.avg_response_ms)}ms
              </span>
            </div>
            {agent.channels.length > 0 && (
              <div className="flex items-center gap-1.5 ml-auto">
                {agent.channels.map((ch) => (
                  <span
                    key={ch}
                    className="text-[#a0a0b8]/40 text-xs font-mono bg-[#0f3460]/50 px-2 py-0.5 rounded"
                  >
                    #{ch}
                  </span>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Tabs */}
        <div className="flex items-center gap-1 mt-3">
          {(["all", "llm", "tools"] as const).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`text-xs font-mono px-3 py-1.5 rounded transition-colors cursor-pointer ${
                tab === t
                  ? "text-[#e2b714] bg-[#e2b714]/10 border border-[#e2b714]/20"
                  : "text-[#a0a0b8]/50 hover:text-[#a0a0b8] border border-transparent"
              }`}
            >
              {t === "all"
                ? "all activity"
                : t === "llm"
                  ? "llm calls"
                  : "tool executions"}
            </button>
          ))}
        </div>
      </div>

      {/* Activity feed */}
      <div className="flex-1 overflow-y-auto p-6">
        {loading && activity.length === 0 ? (
          <div className="text-[#a0a0b8]/50 text-sm font-mono text-center py-12">
            loading...
          </div>
        ) : activity.length === 0 ? (
          <div className="text-[#a0a0b8]/50 text-sm font-mono text-center py-12">
            no activity recorded yet
          </div>
        ) : (
          <div className="max-w-4xl mx-auto space-y-2">
            {activity.map((entry) =>
              entry.kind === "llm" ? (
                <LLMCallEntry key={`llm-${entry.data.id}`} call={entry.data} />
              ) : (
                <ToolExecEntry
                  key={`tool-${entry.data.id}`}
                  exec={entry.data}
                />
              ),
            )}

            {canLoadMore && (
              <button
                onClick={handleLoadMore}
                className="w-full text-center text-[#a0a0b8]/50 hover:text-[#a0a0b8] text-sm font-mono py-3 border border-[#e2b714]/10 rounded-lg hover:border-[#e2b714]/25 transition-colors cursor-pointer"
              >
                load more
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
