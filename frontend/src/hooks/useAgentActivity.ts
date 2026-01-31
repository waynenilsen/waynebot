import { useCallback, useState } from "react";
import * as api from "../api";
import type { AgentStatsResponse, LLMCall, ToolExecution } from "../types";
import { useErrors } from "../store/ErrorContext";

const PAGE_SIZE = 50;

interface UseAgentActivity {
  llmCalls: LLMCall[];
  toolExecs: ToolExecution[];
  stats: AgentStatsResponse | null;
  loading: boolean;
  hasMoreCalls: boolean;
  hasMoreExecs: boolean;
  loadActivity: (personaId: number) => Promise<void>;
  loadMoreCalls: (personaId: number) => Promise<void>;
  loadMoreExecs: (personaId: number) => Promise<void>;
  refreshStats: (personaId: number) => Promise<void>;
  prependLLMCall: (call: LLMCall) => void;
  prependToolExec: (exec: ToolExecution) => void;
}

export function useAgentActivity(): UseAgentActivity {
  const [llmCalls, setLLMCalls] = useState<LLMCall[]>([]);
  const [toolExecs, setToolExecs] = useState<ToolExecution[]>([]);
  const [stats, setStats] = useState<AgentStatsResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [hasMoreCalls, setHasMoreCalls] = useState(false);
  const [hasMoreExecs, setHasMoreExecs] = useState(false);
  const { pushError } = useErrors();

  const loadActivity = useCallback(
    async (personaId: number) => {
      setLoading(true);
      try {
        const [calls, execs, s] = await Promise.all([
          api.getAgentLLMCalls(personaId, { limit: PAGE_SIZE }),
          api.getAgentToolExecutions(personaId, { limit: PAGE_SIZE }),
          api.getAgentStats(personaId),
        ]);
        setLLMCalls(calls);
        setToolExecs(execs);
        setStats(s);
        setHasMoreCalls(calls.length >= PAGE_SIZE);
        setHasMoreExecs(execs.length >= PAGE_SIZE);
      } catch (err) {
        pushError(
          `Failed to load activity: ${err instanceof Error ? err.message : "unknown error"}`,
        );
      } finally {
        setLoading(false);
      }
    },
    [pushError],
  );

  const loadMoreCalls = useCallback(
    async (personaId: number) => {
      try {
        const more = await api.getAgentLLMCalls(personaId, {
          limit: PAGE_SIZE,
          offset: llmCalls.length,
        });
        setLLMCalls((prev) => [...prev, ...more]);
        setHasMoreCalls(more.length >= PAGE_SIZE);
      } catch (err) {
        pushError(
          `Failed to load more calls: ${err instanceof Error ? err.message : "unknown error"}`,
        );
      }
    },
    [llmCalls.length, pushError],
  );

  const loadMoreExecs = useCallback(
    async (personaId: number) => {
      try {
        const more = await api.getAgentToolExecutions(personaId, {
          limit: PAGE_SIZE,
          offset: toolExecs.length,
        });
        setToolExecs((prev) => [...prev, ...more]);
        setHasMoreExecs(more.length >= PAGE_SIZE);
      } catch (err) {
        pushError(
          `Failed to load more executions: ${err instanceof Error ? err.message : "unknown error"}`,
        );
      }
    },
    [toolExecs.length, pushError],
  );

  const refreshStats = useCallback(async (personaId: number) => {
    try {
      const s = await api.getAgentStats(personaId);
      setStats(s);
    } catch {
      // stats refresh is non-critical
    }
  }, []);

  const prependLLMCall = useCallback((call: LLMCall) => {
    setLLMCalls((prev) => [call, ...prev]);
  }, []);

  const prependToolExec = useCallback((exec: ToolExecution) => {
    setToolExecs((prev) => [exec, ...prev]);
  }, []);

  return {
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
  };
}
