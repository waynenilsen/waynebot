import { renderHook, act, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { createElement } from "react";
import type { ReactNode } from "react";
import { AppProvider } from "../store/AppContext";
import { useChannels } from "./useChannels";
import type { Channel } from "../types";

vi.mock("../api", () => ({
  getChannels: vi.fn(),
  createChannel: vi.fn(),
}));

import * as api from "../api";
const mockApi = vi.mocked(api);

const general: Channel = {
  id: 1,
  name: "general",
  description: "General chat",
  created_at: "2024-01-01T00:00:00Z",
};

const random: Channel = {
  id: 2,
  name: "random",
  description: "Random stuff",
  created_at: "2024-01-01T00:00:00Z",
};

function wrapper({ children }: { children: ReactNode }) {
  return createElement(AppProvider, null, children);
}

function scenario() {
  const result = renderHook(() => useChannels(), { wrapper });
  return result;
}

beforeEach(() => {
  vi.resetAllMocks();
  mockApi.getChannels.mockResolvedValue([general, random]);
});

describe("useChannels", () => {
  it("fetches channels on mount", async () => {
    const { result } = scenario();

    await waitFor(() => expect(result.current.channels).toHaveLength(2));
    expect(result.current.channels[0].name).toBe("general");
    expect(result.current.channels[1].name).toBe("random");
    expect(mockApi.getChannels).toHaveBeenCalledOnce();
  });

  it("starts with no current channel", () => {
    const { result } = scenario();
    expect(result.current.currentChannel).toBeNull();
    expect(result.current.currentChannelId).toBeNull();
  });

  it("selects a channel", async () => {
    const { result } = scenario();

    await waitFor(() => expect(result.current.channels).toHaveLength(2));

    act(() => result.current.selectChannel(1));

    expect(result.current.currentChannelId).toBe(1);
    expect(result.current.currentChannel?.name).toBe("general");
  });

  it("creates a channel and adds to list", async () => {
    const newChannel: Channel = {
      id: 3,
      name: "new-channel",
      description: "A new channel",
      created_at: "2024-01-01T00:00:00Z",
    };
    mockApi.createChannel.mockResolvedValue(newChannel);

    const { result } = scenario();

    await waitFor(() => expect(result.current.channels).toHaveLength(2));

    await act(() => result.current.createChannel("new-channel", "A new channel"));

    expect(mockApi.createChannel).toHaveBeenCalledWith(
      "new-channel",
      "A new channel",
    );
    expect(result.current.channels).toHaveLength(3);
    expect(result.current.channels[2].name).toBe("new-channel");
  });

  it("handles fetch failure gracefully", async () => {
    mockApi.getChannels.mockRejectedValue(new Error("Network error"));

    const { result } = scenario();

    // Should not throw, channels stays empty
    await waitFor(() => expect(mockApi.getChannels).toHaveBeenCalled());
    expect(result.current.channels).toEqual([]);
  });
});
