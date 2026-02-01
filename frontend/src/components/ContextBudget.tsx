import { useCallback, useEffect, useState } from "react";
import * as api from "../api";
import type { ContextBudget as ContextBudgetType } from "../types";

interface ContextBudgetProps {
  personaId: number;
  channelId: number;
}

function formatTokens(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(0)}k`;
  return String(n);
}

export default function ContextBudgetBar({
  personaId,
  channelId,
}: ContextBudgetProps) {
  const [budget, setBudget] = useState<ContextBudgetType | null>(null);
  const [resetting, setResetting] = useState(false);

  const fetchBudget = useCallback(async () => {
    try {
      const b = await api.getContextBudget(personaId, channelId);
      setBudget(b);
    } catch {
      // ignore errors silently
    }
  }, [personaId, channelId]);

  useEffect(() => {
    fetchBudget();
  }, [fetchBudget]);

  // Listen for WebSocket context budget updates.
  useEffect(() => {
    const handler = (e: Event) => {
      const detail = (e as CustomEvent).detail;
      if (
        detail &&
        detail.persona_id === personaId &&
        detail.channel_id === channelId
      ) {
        setBudget(detail as ContextBudgetType);
      }
    };
    window.addEventListener("agent_context_budget", handler);
    return () => window.removeEventListener("agent_context_budget", handler);
  }, [personaId, channelId]);

  const handleReset = useCallback(async () => {
    setResetting(true);
    try {
      await api.resetContext(personaId, channelId);
      await fetchBudget();
    } catch {
      // ignore
    } finally {
      setResetting(false);
    }
  }, [personaId, channelId, fetchBudget]);

  if (!budget) return null;

  const used =
    budget.system_tokens +
    budget.project_tokens +
    budget.history_tokens;
  const pct = budget.total_tokens > 0 ? (used / budget.total_tokens) * 100 : 0;
  const warning = pct > 80;
  const full = budget.exhausted;

  const barColor = full
    ? "bg-red-500"
    : warning
      ? "bg-yellow-500"
      : "bg-[#e2b714]/60";

  const textColor = full
    ? "text-red-400"
    : warning
      ? "text-yellow-400"
      : "text-[#a0a0b8]/50";

  return (
    <div className="flex items-center gap-2 px-3 py-1.5 border-t border-[#e2b714]/10 bg-[#1a1a2e]/80">
      <span className={`text-[10px] font-mono ${textColor} shrink-0`}>
        {full ? "context full" : "context"}
      </span>
      <div className="flex-1 h-1 bg-[#0f3460]/50 rounded-full overflow-hidden">
        <div
          className={`h-full ${barColor} rounded-full transition-all`}
          style={{ width: `${Math.min(pct, 100)}%` }}
        />
      </div>
      <span className={`text-[10px] font-mono ${textColor} shrink-0`}>
        {formatTokens(used)}/{formatTokens(budget.total_tokens)}
      </span>
      {full && (
        <button
          onClick={handleReset}
          disabled={resetting}
          className="text-[10px] font-mono text-red-400 hover:text-red-300 border border-red-400/30 hover:border-red-400/50 px-1.5 py-0.5 rounded transition-colors disabled:opacity-50"
        >
          {resetting ? "..." : "reset"}
        </button>
      )}
    </div>
  );
}
