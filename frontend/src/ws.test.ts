import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { connectWs } from "./ws";
import type { ConnectionState } from "./ws";
import type { WsEvent } from "./types";

vi.mock("./utils/token", () => ({
  getToken: vi.fn(() => "tok_test"),
  setToken: vi.fn(),
  clearToken: vi.fn(),
}));

class MockWebSocket {
  static instances: MockWebSocket[] = [];
  onopen: (() => void) | null = null;
  onmessage: ((e: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;
  url: string;
  closed = false;

  constructor(url: string) {
    this.url = url;
    MockWebSocket.instances.push(this);
  }

  close() {
    this.closed = true;
    this.onclose?.();
  }

  simulateOpen() {
    this.onopen?.();
  }

  simulateMessage(data: WsEvent) {
    this.onmessage?.({ data: JSON.stringify(data) });
  }
}

describe("ws", () => {
  beforeEach(() => {
    MockWebSocket.instances = [];
    vi.stubGlobal("WebSocket", MockWebSocket);
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ ticket: "tkt_abc" }),
      }),
    );
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  it("fetches ticket then connects websocket with correct url", async () => {
    const onEvent = vi.fn();
    connectWs(onEvent);

    await vi.advanceTimersByTimeAsync(0);

    expect(fetch).toHaveBeenCalledWith("/api/ws/ticket", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: "Bearer tok_test",
      },
    });

    expect(MockWebSocket.instances).toHaveLength(1);
    expect(MockWebSocket.instances[0].url).toContain("ticket=tkt_abc");
  });

  it("dispatches events to onEvent callback", async () => {
    const onEvent = vi.fn();
    connectWs(onEvent);

    await vi.advanceTimersByTimeAsync(0);
    const ws = MockWebSocket.instances[0];
    ws.simulateOpen();

    const event: WsEvent = {
      type: "new_message",
      data: { id: 1, content: "hello" },
    };
    ws.simulateMessage(event);

    expect(onEvent).toHaveBeenCalledWith(event);
  });

  it("reports connection state changes", async () => {
    const states: ConnectionState[] = [];
    connectWs(vi.fn(), (s) => states.push(s));

    await vi.advanceTimersByTimeAsync(0);
    const ws = MockWebSocket.instances[0];
    ws.simulateOpen();

    expect(states).toContain("connecting");
    expect(states).toContain("connected");
  });

  it("close() tears down the connection", async () => {
    const conn = connectWs(vi.fn());

    await vi.advanceTimersByTimeAsync(0);
    const ws = MockWebSocket.instances[0];
    ws.simulateOpen();

    conn.close();
    expect(ws.closed).toBe(true);
    expect(conn.getState()).toBe("disconnected");
  });

  it("reconnects with exponential backoff on disconnect", async () => {
    connectWs(vi.fn());

    await vi.advanceTimersByTimeAsync(0);
    const ws = MockWebSocket.instances[0];
    ws.simulateOpen();

    // simulate unexpected close
    ws.onclose?.();

    // should reconnect after 1s (initial delay)
    expect(MockWebSocket.instances).toHaveLength(1);
    await vi.advanceTimersByTimeAsync(1000);
    expect(MockWebSocket.instances).toHaveLength(2);
  });
});
