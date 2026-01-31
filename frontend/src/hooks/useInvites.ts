import { useCallback, useEffect, useState } from "react";
import * as api from "../api";
import type { Invite } from "../types";

interface UseInvites {
  invites: Invite[];
  loading: boolean;
  createInvite: () => Promise<Invite>;
  refresh: () => Promise<void>;
}

export function useInvites(): UseInvites {
  const [invites, setInvites] = useState<Invite[]>([]);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.getInvites();
      setInvites(data);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const createInvite = useCallback(async () => {
    const invite = await api.createInvite();
    setInvites((prev) => [invite, ...prev]);
    return invite;
  }, []);

  return { invites, loading, createInvite, refresh };
}
