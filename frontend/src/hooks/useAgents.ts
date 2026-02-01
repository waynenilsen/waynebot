import { useCallback, useState } from "react";
import * as api from "../api";
import type { AgentStatus } from "../types";
import { useErrors } from "../store/ErrorContext";
import { getErrorMessage } from "../utils/errors";

interface UseAgents {
  agents: AgentStatus[];
  supervisorRunning: boolean;
  loading: boolean;
  startAgents: () => Promise<void>;
  stopAgents: () => Promise<void>;
  refresh: () => Promise<void>;
}

export function useAgents(): UseAgents {
  const [agents, setAgents] = useState<AgentStatus[]>([]);
  const [supervisorRunning, setSupervisorRunning] = useState(false);
  const [loading, setLoading] = useState(false);
  const { pushError } = useErrors();

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.getAgentStatus();
      setAgents(data.agents);
      setSupervisorRunning(data.supervisor_running);
    } catch (err) {
      pushError(
        `Failed to load agents: ${getErrorMessage(err)}`,
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
        `Failed to start agents: ${getErrorMessage(err)}`,
      );
    }
  }, [refresh, pushError]);

  const stopAgents = useCallback(async () => {
    try {
      await api.stopAgents();
      await refresh();
    } catch (err) {
      pushError(
        `Failed to stop agents: ${getErrorMessage(err)}`,
      );
    }
  }, [refresh, pushError]);

  return { agents, supervisorRunning, loading, startAgents, stopAgents, refresh };
}
