import { useEffect, useRef, useState } from "react";
import { connectWs } from "../ws";
import type { ConnectionState } from "../ws";
import type { Message, WsEvent } from "../types";
import { useApp } from "../store/AppContext";

export function useWebSocket(authenticated: boolean) {
  const { addMessage } = useApp();
  const [connected, setConnected] = useState(false);
  const connRef = useRef<ReturnType<typeof connectWs> | null>(null);

  useEffect(() => {
    if (!authenticated) {
      connRef.current?.close();
      connRef.current = null;
      setConnected(false);
      return;
    }

    const conn = connectWs(
      (event: WsEvent) => {
        if (event.type === "new_message") {
          addMessage(event.data as Message);
        }
      },
      (state: ConnectionState) => {
        setConnected(state === "connected");
      },
    );
    connRef.current = conn;

    return () => {
      conn.close();
    };
  }, [authenticated, addMessage]);

  return { connected };
}
