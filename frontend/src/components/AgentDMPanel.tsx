import { useEffect, useState, useRef, useCallback } from "react";
import * as api from "../api";
import { useApp } from "../store/AppContext";
import { useMessages } from "../hooks/useMessages";
import { useErrors } from "../store/ErrorContext";
import MessageItem from "./MessageItem";
import MessageCompose from "./MessageCompose";

interface AgentDMPanelProps {
  personaId: number;
  personaName: string;
}

export default function AgentDMPanel({
  personaId,
  personaName,
}: AgentDMPanelProps) {
  const { state, setDMs } = useApp();
  const { pushError } = useErrors();
  const [channelId, setChannelId] = useState<number | null>(null);
  const [initError, setInitError] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);
  const isNearBottomRef = useRef(true);
  const prevCountRef = useRef(0);

  // Create or fetch the DM channel for this persona
  useEffect(() => {
    let cancelled = false;
    setChannelId(null);
    setInitError(false);

    api
      .createDM({ persona_id: personaId })
      .then((dm) => {
        if (cancelled) return;
        setChannelId(dm.id);
        // Keep DM list in sync
        const exists = state.dms.some((d) => d.id === dm.id);
        if (!exists) {
          setDMs([...state.dms, dm]);
        }
      })
      .catch((err) => {
        if (cancelled) return;
        setInitError(true);
        pushError(`Failed to open DM with agent: ${err.message}`);
      });

    return () => {
      cancelled = true;
    };
    // Only re-run when personaId changes
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [personaId]);

  const { messages, loading, hasMore, loadMore, sendMessage, toggleReaction } =
    useMessages(channelId);

  const checkNearBottom = useCallback(() => {
    const el = scrollRef.current;
    if (!el) return;
    isNearBottomRef.current =
      el.scrollHeight - el.scrollTop - el.clientHeight < 80;
  }, []);

  // Auto-scroll on new messages
  useEffect(() => {
    const el = scrollRef.current;
    if (!el) return;

    const grew = messages.length > prevCountRef.current;
    prevCountRef.current = messages.length;

    if (grew && isNearBottomRef.current) {
      el.scrollTop = el.scrollHeight;
    }
  }, [messages]);

  // Scroll to bottom on channel init
  useEffect(() => {
    if (!channelId) return;
    const el = scrollRef.current;
    if (!el) return;
    // Small delay so DOM has rendered messages
    requestAnimationFrame(() => {
      el.scrollTop = el.scrollHeight;
      isNearBottomRef.current = true;
      prevCountRef.current = messages.length;
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [channelId]);

  if (initError) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <p className="text-red-400/60 text-xs font-mono">
          failed to open DM channel
        </p>
      </div>
    );
  }

  if (!channelId) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <span className="text-[#a0a0b8]/40 text-sm font-mono animate-pulse">
          opening dm...
        </span>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col min-h-0">
      {/* Header */}
      <div className="shrink-0 px-4 py-3 border-b border-[#e2b714]/10">
        <h2 className="text-white text-sm font-bold font-mono flex items-center gap-2">
          <span className="text-[#e2b714]/50">●</span>
          {personaName}
        </h2>
      </div>

      {/* Messages */}
      <div
        ref={scrollRef}
        onScroll={checkNearBottom}
        className="flex-1 overflow-y-auto"
      >
        {hasMore && (
          <div className="flex justify-center py-3">
            <button
              onClick={loadMore}
              disabled={loading}
              className="text-xs font-mono text-[#a0a0b8]/60 hover:text-[#e2b714] transition-colors disabled:opacity-40 disabled:cursor-default px-3 py-1 rounded border border-[#a0a0b8]/10 hover:border-[#e2b714]/20"
            >
              {loading ? "loading..." : "load older messages"}
            </button>
          </div>
        )}

        {loading && messages.length === 0 && (
          <div className="flex items-center justify-center py-12">
            <span className="text-[#a0a0b8]/40 text-sm font-mono animate-pulse">
              loading...
            </span>
          </div>
        )}

        {!loading && messages.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 gap-2">
            <div className="text-[#e2b714]/20 text-3xl">~</div>
            <p className="text-[#a0a0b8]/40 text-sm font-mono">
              send a message to this agent
            </p>
          </div>
        )}

        {messages.map((msg) => (
          <MessageItem
            key={msg.id}
            message={msg}
            onReactionToggle={toggleReaction}
          />
        ))}
      </div>

      {/* Compose */}
      <MessageCompose onSend={sendMessage} />
    </div>
  );
}
