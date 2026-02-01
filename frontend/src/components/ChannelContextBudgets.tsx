import { useEffect, useState } from "react";
import * as api from "../api";
import type { ChannelMember } from "../types";
import ContextBudgetBar from "./ContextBudget";

interface ChannelContextBudgetsProps {
  channelId: number;
}

export default function ChannelContextBudgets({
  channelId,
}: ChannelContextBudgetsProps) {
  const [agentMembers, setAgentMembers] = useState<ChannelMember[]>([]);

  useEffect(() => {
    let cancelled = false;
    api
      .getChannelMembers(channelId)
      .then((members) => {
        if (!cancelled) {
          setAgentMembers(members.filter((m) => m.type === "persona"));
        }
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [channelId]);

  if (agentMembers.length === 0) return null;

  return (
    <>
      {agentMembers.map((m) => (
        <ContextBudgetBar
          key={m.id}
          personaId={m.id}
          channelId={channelId}
        />
      ))}
    </>
  );
}
