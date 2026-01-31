import { useEffect, useRef, useCallback } from "react";
import type { Message } from "../types";
import MessageItem from "./MessageItem";

interface MessageThreadProps {
  messages: Message[];
  loading: boolean;
  hasMore: boolean;
  onLoadMore: () => void;
  channelName: string;
}

export default function MessageThread({
  messages,
  loading,
  hasMore,
  onLoadMore,
  channelName,
}: MessageThreadProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const isNearBottomRef = useRef(true);
  const prevMessageCountRef = useRef(0);

  const checkNearBottom = useCallback(() => {
    const el = scrollRef.current;
    if (!el) return;
    const threshold = 80;
    isNearBottomRef.current =
      el.scrollHeight - el.scrollTop - el.clientHeight < threshold;
  }, []);

  // Auto-scroll to bottom when new messages arrive (if user hasn't scrolled up)
  useEffect(() => {
    const el = scrollRef.current;
    if (!el) return;

    const grew = messages.length > prevMessageCountRef.current;
    prevMessageCountRef.current = messages.length;

    if (grew && isNearBottomRef.current) {
      el.scrollTop = el.scrollHeight;
    }
  }, [messages]);

  // Scroll to bottom on initial load / channel switch
  useEffect(() => {
    const el = scrollRef.current;
    if (!el) return;
    el.scrollTop = el.scrollHeight;
    isNearBottomRef.current = true;
    prevMessageCountRef.current = messages.length;
    // Only on channelName change (channel switch)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [channelName]);

  return (
    <div className="flex-1 flex flex-col min-h-0">
      {/* Channel header */}
      <div className="shrink-0 px-4 py-3 border-b border-[#e2b714]/10 bg-[#1a1a2e]">
        <h2 className="text-white text-sm font-bold font-mono flex items-center gap-2">
          <span className="text-[#e2b714]/50">#</span>
          {channelName}
        </h2>
      </div>

      {/* Message list */}
      <div
        ref={scrollRef}
        onScroll={checkNearBottom}
        className="flex-1 overflow-y-auto"
        data-testid="message-scroll"
      >
        {/* Load more */}
        {hasMore && (
          <div className="flex justify-center py-3">
            <button
              onClick={onLoadMore}
              disabled={loading}
              className="text-xs font-mono text-[#a0a0b8]/60 hover:text-[#e2b714] transition-colors disabled:opacity-40 disabled:cursor-default px-3 py-1 rounded border border-[#a0a0b8]/10 hover:border-[#e2b714]/20"
            >
              {loading ? "loading..." : "load older messages"}
            </button>
          </div>
        )}

        {/* Loading spinner for initial load */}
        {loading && messages.length === 0 && (
          <div className="flex-1 flex items-center justify-center py-16">
            <span className="text-[#a0a0b8]/40 text-sm font-mono animate-pulse">
              loading...
            </span>
          </div>
        )}

        {/* Empty state */}
        {!loading && messages.length === 0 && (
          <div className="flex-1 flex flex-col items-center justify-center py-16 gap-2">
            <div className="text-[#e2b714]/20 text-3xl">~</div>
            <p className="text-[#a0a0b8]/40 text-sm font-mono">
              no messages yet — say something
            </p>
          </div>
        )}

        {/* Messages */}
        {messages.map((msg) => (
          <MessageItem key={msg.id} message={msg} />
        ))}
      </div>
    </div>
  );
}
