import { renderHook, act, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { createElement } from "react";
import type { ReactNode } from "react";
import { ErrorProvider } from "../store/ErrorContext";
import { usePersonas } from "./usePersonas";
import type { Persona } from "../types";

vi.mock("../api", () => ({
  getPersonas: vi.fn(),
  createPersona: vi.fn(),
  updatePersona: vi.fn(),
  deletePersona: vi.fn(),
}));

import * as api from "../api";
const mockApi = vi.mocked(api);

const alice: Persona = {
  id: 1,
  name: "alice",
  system_prompt: "You are Alice",
  model: "gpt-4",
  tools_enabled: ["web_search"],
  temperature: 0.7,
  max_tokens: 4096,
  cooldown_secs: 5,
  max_tokens_per_hour: 100000,
  created_at: "2024-01-01T00:00:00Z",
};

const bob: Persona = {
  id: 2,
  name: "bob",
  system_prompt: "You are Bob",
  model: "claude-3-opus",
  tools_enabled: [],
  temperature: 1.0,
  max_tokens: 8192,
  cooldown_secs: 0,
  max_tokens_per_hour: 50000,
  created_at: "2024-01-01T00:00:00Z",
};

function wrapper({ children }: { children: ReactNode }) {
  return createElement(ErrorProvider, null, children);
}

function scenario() {
  return renderHook(() => usePersonas(), { wrapper });
}

beforeEach(() => {
  vi.resetAllMocks();
  mockApi.getPersonas.mockResolvedValue([alice, bob]);
});

describe("usePersonas", () => {
  it("fetches personas on mount", async () => {
    const { result } = scenario();

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.personas).toHaveLength(2);
    expect(result.current.personas[0].name).toBe("alice");
    expect(mockApi.getPersonas).toHaveBeenCalledOnce();
  });

  it("starts in loading state", () => {
    const { result } = scenario();
    expect(result.current.loading).toBe(true);
  });

  it("creates a persona and adds to list", async () => {
    const newPersona: Persona = {
      id: 3,
      name: "charlie",
      system_prompt: "You are Charlie",
      model: "gpt-4",
      tools_enabled: [],
      temperature: 0.5,
      max_tokens: 2048,
      cooldown_secs: 10,
      max_tokens_per_hour: 80000,
      created_at: "2024-01-02T00:00:00Z",
    };
    mockApi.createPersona.mockResolvedValue(newPersona);

    const { result } = scenario();
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(() =>
      result.current.createPersona({
        name: "charlie",
        system_prompt: "You are Charlie",
        model: "gpt-4",
        tools_enabled: [],
        temperature: 0.5,
        max_tokens: 2048,
        cooldown_secs: 10,
        max_tokens_per_hour: 80000,
      }),
    );

    expect(mockApi.createPersona).toHaveBeenCalledOnce();
    expect(result.current.personas).toHaveLength(3);
    expect(result.current.personas[2].name).toBe("charlie");
  });

  it("updates a persona in the list", async () => {
    const updated: Persona = { ...alice, name: "alice-v2" };
    mockApi.updatePersona.mockResolvedValue(updated);

    const { result } = scenario();
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(() =>
      result.current.updatePersona(1, {
        name: "alice-v2",
        system_prompt: alice.system_prompt,
        model: alice.model,
        tools_enabled: alice.tools_enabled,
        temperature: alice.temperature,
        max_tokens: alice.max_tokens,
        cooldown_secs: alice.cooldown_secs,
        max_tokens_per_hour: alice.max_tokens_per_hour,
      }),
    );

    expect(mockApi.updatePersona).toHaveBeenCalledWith(1, expect.anything());
    expect(result.current.personas[0].name).toBe("alice-v2");
  });

  it("deletes a persona from the list", async () => {
    mockApi.deletePersona.mockResolvedValue(undefined);

    const { result } = scenario();
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(() => result.current.deletePersona(1));

    expect(mockApi.deletePersona).toHaveBeenCalledWith(1);
    expect(result.current.personas).toHaveLength(1);
    expect(result.current.personas[0].name).toBe("bob");
  });

  it("handles fetch failure gracefully", async () => {
    mockApi.getPersonas.mockRejectedValue(new Error("Network error"));

    const { result } = scenario();
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.personas).toEqual([]);
  });
});
