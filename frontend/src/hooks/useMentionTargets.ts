import { useCallback, useEffect, useState } from "react";
import * as api from "../api";
import type { MentionTarget } from "../types";

interface UseMentionTargets {
  targets: MentionTarget[];
  loading: boolean;
  refresh: () => Promise<void>;
}

export function useMentionTargets(): UseMentionTargets {
  const [targets, setTargets] = useState<MentionTarget[]>([]);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.getMentionTargets();
      setTargets(data);
    } catch {
      // silently fail - mention autocomplete is non-critical
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return { targets, loading, refresh };
}
