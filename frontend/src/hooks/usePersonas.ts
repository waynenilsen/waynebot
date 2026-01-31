import { useCallback, useEffect, useState } from "react";
import * as api from "../api";
import type { Persona } from "../types";
import { useErrors } from "../store/ErrorContext";

type PersonaData = Omit<Persona, "id" | "created_at">;

interface UsePersonas {
  personas: Persona[];
  loading: boolean;
  createPersona: (data: PersonaData) => Promise<Persona>;
  updatePersona: (id: number, data: PersonaData) => Promise<Persona>;
  deletePersona: (id: number) => Promise<void>;
  refresh: () => Promise<void>;
}

export function usePersonas(): UsePersonas {
  const [personas, setPersonas] = useState<Persona[]>([]);
  const [loading, setLoading] = useState(true);
  const { pushError } = useErrors();

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.getPersonas();
      setPersonas(data);
    } catch (err) {
      pushError(
        `Failed to load personas: ${err instanceof Error ? err.message : "unknown error"}`,
      );
    } finally {
      setLoading(false);
    }
  }, [pushError]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const createPersona = useCallback(
    async (data: PersonaData) => {
      try {
        const persona = await api.createPersona(data);
        setPersonas((prev) => [...prev, persona]);
        return persona;
      } catch (err) {
        pushError(
          `Failed to create persona: ${err instanceof Error ? err.message : "unknown error"}`,
        );
        throw err;
      }
    },
    [pushError],
  );

  const updatePersona = useCallback(
    async (id: number, data: PersonaData) => {
      try {
        const persona = await api.updatePersona(id, data);
        setPersonas((prev) => prev.map((p) => (p.id === id ? persona : p)));
        return persona;
      } catch (err) {
        pushError(
          `Failed to update persona: ${err instanceof Error ? err.message : "unknown error"}`,
        );
        throw err;
      }
    },
    [pushError],
  );

  const deletePersona = useCallback(
    async (id: number) => {
      try {
        await api.deletePersona(id);
        setPersonas((prev) => prev.filter((p) => p.id !== id));
      } catch (err) {
        pushError(
          `Failed to delete persona: ${err instanceof Error ? err.message : "unknown error"}`,
        );
        throw err;
      }
    },
    [pushError],
  );

  return {
    personas,
    loading,
    createPersona,
    updatePersona,
    deletePersona,
    refresh,
  };
}
