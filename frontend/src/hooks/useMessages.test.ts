import { renderHook, act, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { createElement } from "react";
import type { ReactNode } from "react";
import { AppProvider } from "../store/AppContext";
import { useMessages } from "./useMessages";
import type { Message } from "../types";

vi.mock("../api", () => ({
  getMessages: vi.fn(),
  postMessage: vi.fn(),
}));

import * as api from "../api";
const mockApi = vi.mocked(api);

function msg(id: number, content = `message ${id}`): Message {
  return {
    id,
    channel_id: 1,
    author_id: 1,
    author_type: "human",
    author_name: "alice",
    content,
    created_at: "2024-01-01T00:00:00Z",
  };
}

function wrapper({ children }: { children: ReactNode }) {
  return createElement(AppProvider, null, children);
}

function scenario(channelId: number | null = 1) {
  return renderHook(() => useMessages(channelId), { wrapper });
}

beforeEach(() => {
  vi.resetAllMocks();
  mockApi.getMessages.mockResolvedValue([msg(1), msg(2), msg(3)]);
});

describe("useMessages", () => {
  it("fetches messages for a channel on mount", async () => {
    const { result } = scenario(1);

    await waitFor(() => expect(result.current.messages).toHaveLength(3));
    expect(mockApi.getMessages).toHaveBeenCalledWith(1, { limit: 50 });
  });

  it("returns empty messages for null channel", () => {
    const { result } = scenario(null);
    expect(result.current.messages).toEqual([]);
    expect(mockApi.getMessages).not.toHaveBeenCalled();
  });

  it("sets hasMore false when fewer than 50 messages returned", async () => {
    mockApi.getMessages.mockResolvedValue([msg(1)]);
    const { result } = scenario(1);

    await waitFor(() => expect(result.current.messages).toHaveLength(1));
    expect(result.current.hasMore).toBe(false);
  });

  it("sets hasMore true when 50 messages returned", async () => {
    const fiftyMsgs = Array.from({ length: 50 }, (_, i) => msg(i + 1));
    mockApi.getMessages.mockResolvedValue(fiftyMsgs);
    const { result } = scenario(1);

    await waitFor(() => expect(result.current.messages).toHaveLength(50));
    expect(result.current.hasMore).toBe(true);
  });

  it("loadMore fetches older messages and prepends them", async () => {
    const fiftyMsgs = Array.from({ length: 50 }, (_, i) => msg(i + 1));
    mockApi.getMessages.mockResolvedValue(fiftyMsgs);
    const { result } = scenario(1);

    await waitFor(() => expect(result.current.messages).toHaveLength(50));

    const olderMsgs = [msg(100, "old1"), msg(101, "old2")];
    mockApi.getMessages.mockResolvedValue(olderMsgs);

    await act(() => result.current.loadMore());

    expect(mockApi.getMessages).toHaveBeenCalledWith(1, {
      limit: 50,
      before: 1,
    });
    expect(result.current.messages).toHaveLength(52);
    expect(result.current.messages[0].content).toBe("old1");
  });

  it("sendMessage posts and adds message to state", async () => {
    const newMsg = msg(99, "hello");
    mockApi.postMessage.mockResolvedValue(newMsg);
    const { result } = scenario(1);

    await waitFor(() => expect(result.current.messages).toHaveLength(3));

    await act(() => result.current.sendMessage("hello"));

    expect(mockApi.postMessage).toHaveBeenCalledWith(1, "hello");
    expect(result.current.messages).toHaveLength(4);
    expect(result.current.messages[3].content).toBe("hello");
  });
});
