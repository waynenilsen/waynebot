import { renderHook, act, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { useAuth } from "./useAuth";
import type { AuthResponse, User } from "../types";

const alice: User = {
  id: 1,
  username: "alice",
  created_at: "2024-01-01T00:00:00Z",
};
const authResp: AuthResponse = { token: "tok_abc", user: alice };

vi.mock("../api", () => ({
  login: vi.fn(),
  register: vi.fn(),
  logout: vi.fn(),
  getMe: vi.fn(),
}));

vi.mock("../utils/token", () => ({
  getToken: vi.fn(),
  setToken: vi.fn(),
  clearToken: vi.fn(),
}));

import * as api from "../api";
import * as tokenUtils from "../utils/token";

const mockApi = vi.mocked(api);
const mockToken = vi.mocked(tokenUtils);

beforeEach(() => {
  vi.resetAllMocks();
  mockToken.getToken.mockReturnValue(null);
});

describe("useAuth", () => {
  it("resolves to not loading when no token exists", async () => {
    const { result } = renderHook(() => useAuth());
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.user).toBeNull();
  });

  it("starts loading when token exists", () => {
    mockToken.getToken.mockReturnValue("tok_abc");
    mockApi.getMe.mockReturnValue(new Promise(() => {})); // never resolves
    const { result } = renderHook(() => useAuth());
    expect(result.current.loading).toBe(true);
    expect(result.current.user).toBeNull();
  });

  it("restores session from token on mount", async () => {
    mockToken.getToken.mockReturnValue("tok_abc");
    mockApi.getMe.mockResolvedValue(alice);

    const { result } = renderHook(() => useAuth());
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(mockApi.getMe).toHaveBeenCalled();
    expect(result.current.user).toEqual(alice);
  });

  it("clears token when session restore fails", async () => {
    mockToken.getToken.mockReturnValue("tok_expired");
    mockApi.getMe.mockRejectedValue(new Error("Unauthorized"));

    const { result } = renderHook(() => useAuth());
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(mockToken.clearToken).toHaveBeenCalled();
    expect(result.current.user).toBeNull();
  });

  it("login sets user on success", async () => {
    mockApi.login.mockResolvedValue(authResp);

    const { result } = renderHook(() => useAuth());
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(() => result.current.login("alice", "password123"));

    expect(mockApi.login).toHaveBeenCalledWith("alice", "password123");
    expect(result.current.user).toEqual(alice);
  });

  it("login propagates errors", async () => {
    mockApi.login.mockRejectedValue(new Error("Invalid credentials"));

    const { result } = renderHook(() => useAuth());
    await waitFor(() => expect(result.current.loading).toBe(false));

    await expect(
      act(() => result.current.login("alice", "wrong")),
    ).rejects.toThrow("Invalid credentials");
    expect(result.current.user).toBeNull();
  });

  it("register sets user on success", async () => {
    mockApi.register.mockResolvedValue(authResp);

    const { result } = renderHook(() => useAuth());
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(() =>
      result.current.register("alice", "password123", "inv_code"),
    );

    expect(mockApi.register).toHaveBeenCalledWith(
      "alice",
      "password123",
      "inv_code",
    );
    expect(result.current.user).toEqual(alice);
  });

  it("register works without invite code", async () => {
    mockApi.register.mockResolvedValue(authResp);

    const { result } = renderHook(() => useAuth());
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(() => result.current.register("alice", "password123"));

    expect(mockApi.register).toHaveBeenCalledWith(
      "alice",
      "password123",
      undefined,
    );
  });

  it("logout clears user", async () => {
    mockToken.getToken.mockReturnValue("tok_abc");
    mockApi.getMe.mockResolvedValue(alice);
    mockApi.logout.mockResolvedValue(undefined);

    const { result } = renderHook(() => useAuth());
    await waitFor(() => expect(result.current.user).toEqual(alice));

    await act(() => result.current.logout());

    expect(mockApi.logout).toHaveBeenCalled();
    expect(result.current.user).toBeNull();
  });
});
