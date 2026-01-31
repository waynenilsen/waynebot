import { useCallback, useEffect, useState } from "react";
import * as api from "../api";
import type { User } from "../types";
import { clearToken, getToken } from "../utils/token";

interface UseAuth {
  user: User | null;
  loading: boolean;
  login: (username: string, password: string) => Promise<void>;
  register: (
    username: string,
    password: string,
    inviteCode?: string,
  ) => Promise<void>;
  logout: () => Promise<void>;
}

export function useAuth(): UseAuth {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = getToken();
    if (!token) {
      setLoading(false);
      return;
    }
    api
      .getMe()
      .then((u) => setUser(u))
      .catch(() => clearToken())
      .finally(() => setLoading(false));
  }, []);

  const login = useCallback(async (username: string, password: string) => {
    const resp = await api.login(username, password);
    setUser(resp.user);
  }, []);

  const register = useCallback(
    async (username: string, password: string, inviteCode?: string) => {
      const resp = await api.register(username, password, inviteCode);
      setUser(resp.user);
    },
    [],
  );

  const logout = useCallback(async () => {
    await api.logout();
    setUser(null);
  }, []);

  return { user, loading, login, register, logout };
}
