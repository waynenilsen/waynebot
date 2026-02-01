import { useCallback, useEffect, useState } from "react";
import * as api from "../api";
import type { Project } from "../types";
import { useErrors } from "../store/ErrorContext";

type ProjectData = Omit<Project, "id" | "created_at">;

interface UseProjects {
  projects: Project[];
  loading: boolean;
  createProject: (data: ProjectData) => Promise<Project>;
  updateProject: (id: number, data: ProjectData) => Promise<Project>;
  deleteProject: (id: number) => Promise<void>;
  refresh: () => Promise<void>;
}

export function useProjects(): UseProjects {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const { pushError } = useErrors();

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.getProjects();
      setProjects(data);
    } catch (err) {
      pushError(
        `Failed to load projects: ${err instanceof Error ? err.message : "unknown error"}`,
      );
    } finally {
      setLoading(false);
    }
  }, [pushError]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const createProject = useCallback(
    async (data: ProjectData) => {
      try {
        const project = await api.createProject(data);
        setProjects((prev) => [...prev, project]);
        return project;
      } catch (err) {
        pushError(
          `Failed to create project: ${err instanceof Error ? err.message : "unknown error"}`,
        );
        throw err;
      }
    },
    [pushError],
  );

  const updateProject = useCallback(
    async (id: number, data: ProjectData) => {
      try {
        const project = await api.updateProject(id, data);
        setProjects((prev) => prev.map((p) => (p.id === id ? project : p)));
        return project;
      } catch (err) {
        pushError(
          `Failed to update project: ${err instanceof Error ? err.message : "unknown error"}`,
        );
        throw err;
      }
    },
    [pushError],
  );

  const deleteProject = useCallback(
    async (id: number) => {
      try {
        await api.deleteProject(id);
        setProjects((prev) => prev.filter((p) => p.id !== id));
      } catch (err) {
        pushError(
          `Failed to delete project: ${err instanceof Error ? err.message : "unknown error"}`,
        );
        throw err;
      }
    },
    [pushError],
  );

  return {
    projects,
    loading,
    createProject,
    updateProject,
    deleteProject,
    refresh,
  };
}

interface UseChannelProjects {
  projects: Project[];
  loading: boolean;
  addProject: (projectId: number) => Promise<void>;
  removeProject: (projectId: number) => Promise<void>;
  refresh: () => Promise<void>;
}

export function useChannelProjects(channelId: number): UseChannelProjects {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const { pushError } = useErrors();

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.getChannelProjects(channelId);
      setProjects(data);
    } catch (err) {
      pushError(
        `Failed to load channel projects: ${err instanceof Error ? err.message : "unknown error"}`,
      );
    } finally {
      setLoading(false);
    }
  }, [channelId, pushError]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const addProject = useCallback(
    async (projectId: number) => {
      try {
        await api.addChannelProject(channelId, projectId);
        await refresh();
      } catch (err) {
        pushError(
          `Failed to add project: ${err instanceof Error ? err.message : "unknown error"}`,
        );
        throw err;
      }
    },
    [channelId, pushError, refresh],
  );

  const removeProject = useCallback(
    async (projectId: number) => {
      try {
        await api.removeChannelProject(channelId, projectId);
        await refresh();
      } catch (err) {
        pushError(
          `Failed to remove project: ${err instanceof Error ? err.message : "unknown error"}`,
        );
        throw err;
      }
    },
    [channelId, pushError, refresh],
  );

  return {
    projects,
    loading,
    addProject,
    removeProject,
    refresh,
  };
}
