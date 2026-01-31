import { useCallback, useState } from "react";
import * as api from "../api";
import type { AgentStatus } from "../types";

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

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.getAgentStatus();
      setAgents(data);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, []);

  const startAgents = useCallback(async () => {
    await api.startAgents();
    await refresh();
  }, [refresh]);

  const stopAgents = useCallback(async () => {
    await api.stopAgents();
    await refresh();
  }, [refresh]);

  return { agents, loading, startAgents, stopAgents, refresh };
}
