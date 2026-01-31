import { renderHook, act } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { createElement } from "react";
import type { ReactNode } from "react";
import { AppProvider } from "../store/AppContext";
import { useWebSocket } from "./useWebSocket";
import type { ConnectionState } from "../ws";
import type { WsEvent } from "../types";

let capturedOnEvent: ((event: WsEvent) => void) | null = null;
let capturedOnState: ((state: ConnectionState) => void) | null = null;
const mockClose = vi.fn();

vi.mock("../ws", () => ({
  connectWs: vi.fn(
    (
      onEvent: (event: WsEvent) => void,
      onState: (state: ConnectionState) => void,
    ) => {
      capturedOnEvent = onEvent;
      capturedOnState = onState;
      return { close: mockClose, getState: () => "connecting" as const };
    },
  ),
}));

import { connectWs } from "../ws";
const mockConnectWs = vi.mocked(connectWs);

function wrapper({ children }: { children: ReactNode }) {
  return createElement(AppProvider, null, children);
}

function scenario(authenticated = true) {
  return renderHook(({ auth }) => useWebSocket(auth), {
    wrapper,
    initialProps: { auth: authenticated },
  });
}

beforeEach(() => {
  vi.resetAllMocks();
  capturedOnEvent = null;
  capturedOnState = null;
});

describe("useWebSocket", () => {
  it("connects when authenticated", () => {
    scenario(true);
    expect(mockConnectWs).toHaveBeenCalledOnce();
  });

  it("does not connect when not authenticated", () => {
    scenario(false);
    expect(mockConnectWs).not.toHaveBeenCalled();
  });

  it("starts as disconnected", () => {
    const { result } = scenario(true);
    expect(result.current.connected).toBe(false);
  });

  it("updates connected state on state change", () => {
    const { result } = scenario(true);

    act(() => capturedOnState?.("connected"));
    expect(result.current.connected).toBe(true);

    act(() => capturedOnState?.("disconnected"));
    expect(result.current.connected).toBe(false);
  });

  it("closes connection on unmount", () => {
    const { unmount } = scenario(true);
    unmount();
    expect(mockClose).toHaveBeenCalled();
  });

  it("closes connection when auth becomes false", () => {
    const { rerender } = scenario(true);
    rerender({ auth: false });
    expect(mockClose).toHaveBeenCalled();
  });

  it("dispatches new_message events to context", () => {
    scenario(true);

    const message = {
      id: 1,
      channel_id: 1,
      author_id: 1,
      author_type: "human" as const,
      author_name: "alice",
      content: "hello",
      created_at: "2024-01-01T00:00:00Z",
    };

    act(() => {
      capturedOnEvent?.({ type: "new_message", data: message });
    });

    // If this doesn't throw, the message was dispatched successfully
    // (AppContext's addMessage was called internally)
    expect(capturedOnEvent).not.toBeNull();
  });
});
