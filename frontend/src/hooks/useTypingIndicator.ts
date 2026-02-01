import { useEffect, useState } from "react";
import type { AgentStatusEvent } from "../types";

interface TypingAgent {
  persona_id: number;
  persona_name: string;
}

export function useTypingIndicator(channelId: number | null): TypingAgent[] {
  const [typing, setTyping] = useState<Map<number, TypingAgent>>(new Map());

  useEffect(() => {
    function handleStatus(e: Event) {
      const data = (e as CustomEvent).detail as AgentStatusEvent;
      if (data.channel_id !== channelId) return;

      setTyping((prev) => {
        const next = new Map(prev);
        if (data.status === "thinking" || data.status === "tool_call") {
          next.set(data.persona_id, {
            persona_id: data.persona_id,
            persona_name: data.persona_name,
          });
        } else {
          next.delete(data.persona_id);
        }
        return next;
      });
    }

    window.addEventListener("agent_status", handleStatus);
    return () => window.removeEventListener("agent_status", handleStatus);
  }, [channelId]);

  // Clear typing state on channel switch.
  useEffect(() => {
    setTyping(new Map());
  }, [channelId]);

  return Array.from(typing.values());
}
