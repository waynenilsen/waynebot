import { useCallback, useEffect, useState } from "react";
import * as api from "../api";
import type { Persona } from "../types";

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

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.getPersonas();
      setPersonas(data);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const createPersona = useCallback(async (data: PersonaData) => {
    const persona = await api.createPersona(data);
    setPersonas((prev) => [...prev, persona]);
    return persona;
  }, []);

  const updatePersona = useCallback(async (id: number, data: PersonaData) => {
    const persona = await api.updatePersona(id, data);
    setPersonas((prev) => prev.map((p) => (p.id === id ? persona : p)));
    return persona;
  }, []);

  const deletePersona = useCallback(async (id: number) => {
    await api.deletePersona(id);
    setPersonas((prev) => prev.filter((p) => p.id !== id));
  }, []);

  return {
    personas,
    loading,
    createPersona,
    updatePersona,
    deletePersona,
    refresh,
  };
}
