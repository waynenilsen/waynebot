import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { login, getChannels, postMessage, logout, ApiError } from "./api";
import * as tokenMod from "./utils/token";

vi.mock("./utils/token", () => ({
  getToken: vi.fn(() => null),
  setToken: vi.fn(),
  clearToken: vi.fn(),
}));

function mockFetch(body: unknown, status = 200) {
  return vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    statusText: "Error",
    json: () => Promise.resolve(body),
  });
}

describe("api", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", mockFetch({}));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("login", () => {
    it("sends correct body and returns parsed response", async () => {
      const authResp = {
        token: "tok_123",
        user: { id: 1, username: "alice", created_at: "2024-01-01T00:00:00Z" },
      };
      vi.stubGlobal("fetch", mockFetch(authResp));

      const result = await login("alice", "password123");

      expect(fetch).toHaveBeenCalledWith("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username: "alice", password: "password123" }),
      });
      expect(result).toEqual(authResp);
      expect(tokenMod.setToken).toHaveBeenCalledWith("tok_123");
    });
  });

  describe("error handling", () => {
    it("throws ApiError with message from response body", async () => {
      vi.stubGlobal(
        "fetch",
        mockFetch({ error: "invalid credentials" }, 401),
      );

      await expect(login("alice", "wrong")).rejects.toThrow(ApiError);
      await expect(login("alice", "wrong")).rejects.toThrow(
        "invalid credentials",
      );
    });

    it("falls back to statusText when body has no error field", async () => {
      vi.stubGlobal("fetch", mockFetch({}, 500));

      await expect(login("alice", "x")).rejects.toThrow("Error");
    });
  });

  describe("getChannels", () => {
    it("sends auth header when token exists", async () => {
      vi.mocked(tokenMod.getToken).mockReturnValue("tok_abc");
      const channels = [
        {
          id: 1,
          name: "general",
          description: "General chat",
          created_at: "2024-01-01T00:00:00Z",
        },
      ];
      vi.stubGlobal("fetch", mockFetch(channels));

      const result = await getChannels();

      expect(fetch).toHaveBeenCalledWith("/api/channels", {
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer tok_abc",
        },
      });
      expect(result).toEqual(channels);
    });
  });

  describe("postMessage", () => {
    it("posts to correct channel endpoint", async () => {
      vi.mocked(tokenMod.getToken).mockReturnValue("tok_abc");
      const msg = {
        id: 1,
        channel_id: 5,
        author_id: 1,
        author_type: "human",
        author_name: "alice",
        content: "hello",
        created_at: "2024-01-01T00:00:00Z",
      };
      vi.stubGlobal("fetch", mockFetch(msg));

      const result = await postMessage(5, "hello");

      expect(fetch).toHaveBeenCalledWith("/api/channels/5/messages", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer tok_abc",
        },
        body: JSON.stringify({ content: "hello" }),
      });
      expect(result).toEqual(msg);
    });
  });

  describe("logout", () => {
    it("clears token after successful call", async () => {
      vi.mocked(tokenMod.getToken).mockReturnValue("tok_abc");
      vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValue({
          ok: true,
          status: 204,
          statusText: "No Content",
          json: () => Promise.resolve(undefined),
        }),
      );

      await logout();

      expect(tokenMod.clearToken).toHaveBeenCalled();
    });
  });
});
