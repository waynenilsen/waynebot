import { useEffect, useRef, useState } from "react";
import { connectWs } from "../ws";
import type { ConnectionState } from "../ws";
import type { Message, WsEvent } from "../types";
import { useApp } from "../store/AppContext";

export function useWebSocket(authenticated: boolean) {
  const { state, addMessage, incrementUnread } = useApp();
  const [connected, setConnected] = useState(false);
  const [wasConnected, setWasConnected] = useState(false);
  const connRef = useRef<ReturnType<typeof connectWs> | null>(null);
  const currentChannelRef = useRef<number | null>(null);
  const userIdRef = useRef<number | null>(null);

  // Keep refs in sync without triggering reconnection.
  currentChannelRef.current = state.currentChannelId;
  userIdRef.current = state.user?.id ?? null;

  useEffect(() => {
    if (!authenticated) {
      connRef.current?.close();
      connRef.current = null;
      setConnected(false);
      return;
    }

    let hadConnection = false;

    const conn = connectWs(
      (event: WsEvent) => {
        if (event.type === "new_message") {
          const msg = event.data as Message;
          addMessage(msg);
          // Increment unread if message is in a different channel than the active one
          // and was not sent by the current user.
          if (
            msg.channel_id !== currentChannelRef.current &&
            msg.author_id !== userIdRef.current
          ) {
            incrementUnread(msg.channel_id);
          }
        } else if (
          event.type === "agent_llm_call" ||
          event.type === "agent_tool_execution"
        ) {
          window.dispatchEvent(
            new CustomEvent(event.type, { detail: event.data }),
          );
        }
      },
      (state: ConnectionState) => {
        const isConnected = state === "connected";
        setConnected(isConnected);
        if (isConnected && hadConnection) {
          setWasConnected(true);
        }
        if (isConnected) {
          hadConnection = true;
        }
      },
    );
    connRef.current = conn;

    return () => {
      conn.close();
    };
  }, [authenticated, addMessage, incrementUnread]);

  return { connected, wasConnected };
}
