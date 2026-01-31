import { useCallback, useState } from "react";
import * as api from "../api";
import type { AgentStatus } from "../types";
import { useErrors } from "../store/ErrorContext";

interface UseAgents {
  agents: AgentStatus[];
  loading: boolean;
  startAgents: () => Promise<void>;
  stopAgents: () => Promise<void>;
  refresh: () => Promise<void>;
}

export function useAgents(): UseAgents {
  const [agents, setAgents] = useState<AgentStatus[]>([]);
  const [loading, setLoading] = useState(false);
  const { pushError } = useErrors();

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.getAgentStatus();
      setAgents(data);
    } catch (err) {
      pushError(
        `Failed to load agents: ${err instanceof Error ? err.message : "unknown error"}`,
      );
    } finally {
      setLoading(false);
    }
  }, [pushError]);

  const startAgents = useCallback(async () => {
    try {
      await api.startAgents();
      await refresh();
    } catch (err) {
      pushError(
        `Failed to start agents: ${err instanceof Error ? err.message : "unknown error"}`,
      );
    }
  }, [refresh, pushError]);

  const stopAgents = useCallback(async () => {
    try {
      await api.stopAgents();
      await refresh();
    } catch (err) {
      pushError(
        `Failed to stop agents: ${err instanceof Error ? err.message : "unknown error"}`,
      );
    }
  }, [refresh, pushError]);

  return { agents, loading, startAgents, stopAgents, refresh };
}
