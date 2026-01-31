import { useCallback, useEffect, useState } from "react";
import * as api from "../api";
import type { Invite } from "../types";
import { useErrors } from "../store/ErrorContext";

interface UseInvites {
  invites: Invite[];
  loading: boolean;
  createInvite: () => Promise<Invite>;
  refresh: () => Promise<void>;
}

export function useInvites(): UseInvites {
  const [invites, setInvites] = useState<Invite[]>([]);
  const [loading, setLoading] = useState(true);
  const { pushError } = useErrors();

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.getInvites();
      setInvites(data);
    } catch (err) {
      pushError(
        `Failed to load invites: ${err instanceof Error ? err.message : "unknown error"}`,
      );
    } finally {
      setLoading(false);
    }
  }, [pushError]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const createInvite = useCallback(async () => {
    try {
      const invite = await api.createInvite();
      setInvites((prev) => [invite, ...prev]);
      return invite;
    } catch (err) {
      pushError(
        `Failed to generate invite: ${err instanceof Error ? err.message : "unknown error"}`,
      );
      throw err;
    }
  }, [pushError]);

  return { invites, loading, createInvite, refresh };
}
