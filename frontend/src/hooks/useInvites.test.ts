import { renderHook, act, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { createElement } from "react";
import type { ReactNode } from "react";
import { ErrorProvider } from "../store/ErrorContext";
import { useInvites } from "./useInvites";
import type { Invite } from "../types";

vi.mock("../api", () => ({
  getInvites: vi.fn(),
  createInvite: vi.fn(),
}));

import * as api from "../api";
const mockApi = vi.mocked(api);

const invite1: Invite = {
  id: 1,
  code: "abc-123-xyz",
  created_by: 1,
  used_by: null,
  created_at: "2024-01-01T00:00:00Z",
};

const invite2: Invite = {
  id: 2,
  code: "def-456-uvw",
  created_by: 1,
  used_by: 5,
  created_at: "2024-01-02T00:00:00Z",
};

function wrapper({ children }: { children: ReactNode }) {
  return createElement(ErrorProvider, null, children);
}

function scenario() {
  return renderHook(() => useInvites(), { wrapper });
}

beforeEach(() => {
  vi.resetAllMocks();
  mockApi.getInvites.mockResolvedValue([invite1, invite2]);
});

describe("useInvites", () => {
  it("fetches invites on mount", async () => {
    const { result } = scenario();

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.invites).toHaveLength(2);
    expect(result.current.invites[0].code).toBe("abc-123-xyz");
    expect(mockApi.getInvites).toHaveBeenCalledOnce();
  });

  it("starts in loading state", () => {
    const { result } = scenario();
    expect(result.current.loading).toBe(true);
  });

  it("creates an invite and prepends to list", async () => {
    const newInvite: Invite = {
      id: 3,
      code: "ghi-789-rst",
      created_by: 1,
      used_by: null,
      created_at: "2024-01-03T00:00:00Z",
    };
    mockApi.createInvite.mockResolvedValue(newInvite);

    const { result } = scenario();
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(() => result.current.createInvite());

    expect(mockApi.createInvite).toHaveBeenCalledOnce();
    expect(result.current.invites).toHaveLength(3);
    expect(result.current.invites[0].code).toBe("ghi-789-rst");
  });

  it("handles fetch failure gracefully", async () => {
    mockApi.getInvites.mockRejectedValue(new Error("Network error"));

    const { result } = scenario();
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.invites).toEqual([]);
  });
});
