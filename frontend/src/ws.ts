import type { WsEvent } from "./types";
import { getToken } from "./utils/token";

export type ConnectionState = "connecting" | "connected" | "disconnected";

export interface WsConnection {
  close: () => void;
  getState: () => ConnectionState;
}

export function connectWs(
  onEvent: (event: WsEvent) => void,
  onStateChange?: (state: ConnectionState) => void,
): WsConnection {
  let state: ConnectionState = "disconnected";
  let ws: WebSocket | null = null;
  let closed = false;
  let retryDelay = 1000;
  let retryTimeout: ReturnType<typeof setTimeout> | null = null;

  function setState(next: ConnectionState) {
    state = next;
    onStateChange?.(next);
  }

  async function connect() {
    if (closed) return;
    setState("connecting");

    try {
      const token = getToken();
      const res = await fetch("/api/ws/ticket", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
      });
      if (!res.ok) throw new Error("Failed to get ws ticket");
      const { ticket } = (await res.json()) as { ticket: string };

      const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
      const url = `${proto}//${window.location.host}/ws?ticket=${ticket}`;

      ws = new WebSocket(url);

      ws.onopen = () => {
        setState("connected");
        retryDelay = 1000;
      };

      ws.onmessage = (e) => {
        try {
          const event = JSON.parse(e.data as string) as WsEvent;
          onEvent(event);
        } catch {
          // ignore malformed messages
        }
      };

      ws.onclose = () => {
        ws = null;
        if (!closed) {
          setState("disconnected");
          scheduleReconnect();
        }
      };

      ws.onerror = () => {
        ws?.close();
      };
    } catch {
      setState("disconnected");
      if (!closed) scheduleReconnect();
    }
  }

  function scheduleReconnect() {
    retryTimeout = setTimeout(() => {
      retryDelay = Math.min(retryDelay * 2, 30000);
      connect();
    }, retryDelay);
  }

  connect();

  return {
    close() {
      closed = true;
      if (retryTimeout) clearTimeout(retryTimeout);
      ws?.close();
      setState("disconnected");
    },
    getState() {
      return state;
    },
  };
}
