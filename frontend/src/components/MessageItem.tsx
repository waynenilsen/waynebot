import { useState } from "react";
import type { Message, ReactionCount } from "../types";
import MarkdownRenderer from "./MarkdownRenderer";

function formatRelativeTime(dateStr: string): string {
  const now = Date.now();
  const then = new Date(dateStr).getTime();
  const diff = now - then;
  const seconds = Math.floor(diff / 1000);

  if (seconds < 60) return "just now";
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d ago`;

  return new Date(dateStr).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
  });
}

function HumanAvatar({ name }: { name: string }) {
  return (
    <div className="w-8 h-8 rounded-full bg-[#0f3460] border border-[#a0a0b8]/15 flex items-center justify-center shrink-0">
      <span className="text-[#a0a0b8] text-xs font-bold uppercase">
        {name.charAt(0)}
      </span>
    </div>
  );
}

function AgentAvatar({ name }: { name: string }) {
  return (
    <div className="w-8 h-8 shrink-0 flex items-center justify-center">
      <div className="w-6 h-6 rotate-45 bg-[#e2b714]/15 border border-[#e2b714]/30 flex items-center justify-center">
        <span className="-rotate-45 text-[#e2b714] text-[10px] font-bold uppercase">
          {name.charAt(0)}
        </span>
      </div>
    </div>
  );
}

const QUICK_EMOJI = [
  "\u{1F44D}",
  "\u{2764}\u{FE0F}",
  "\u{1F604}",
  "\u{1F44E}",
  "\u{1F440}",
  "\u{1F389}",
  "\u{1F914}",
  "\u{1F525}",
];

interface ReactionPillsProps {
  reactions: ReactionCount[] | null;
  onToggle: (emoji: string, currentlyReacted: boolean) => void;
}

function ReactionPills({ reactions, onToggle }: ReactionPillsProps) {
  const [showPicker, setShowPicker] = useState(false);

  const pills = reactions?.filter((r) => r.count > 0) ?? [];

  return (
    <div className="flex flex-wrap items-center gap-1 mt-1">
      {pills.map((r) => (
        <button
          key={r.emoji}
          onClick={() => onToggle(r.emoji, r.reacted)}
          className={`inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-xs border transition-colors ${
            r.reacted
              ? "border-[#e2b714]/40 bg-[#e2b714]/10 text-[#e2b714]"
              : "border-[#a0a0b8]/15 bg-[#a0a0b8]/5 text-[#a0a0b8]/70 hover:border-[#a0a0b8]/30"
          }`}
        >
          <span>{r.emoji}</span>
          <span className="font-mono text-[10px]">{r.count}</span>
        </button>
      ))}

      {/* Add reaction button */}
      <div className="relative">
        <button
          onClick={() => setShowPicker((p) => !p)}
          className="inline-flex items-center px-1.5 py-0.5 rounded text-xs border border-transparent text-[#a0a0b8]/30 hover:text-[#a0a0b8]/60 hover:border-[#a0a0b8]/15 transition-colors opacity-0 group-hover:opacity-100"
          title="Add reaction"
        >
          +
        </button>
        {showPicker && (
          <>
            <div
              className="fixed inset-0 z-40"
              onClick={() => setShowPicker(false)}
            />
            <div className="absolute bottom-full left-0 mb-1 z-50 bg-[#1a1a2e] border border-[#a0a0b8]/15 rounded-lg p-1.5 flex gap-0.5 shadow-lg">
              {QUICK_EMOJI.map((emoji) => (
                <button
                  key={emoji}
                  onClick={() => {
                    onToggle(emoji, false);
                    setShowPicker(false);
                  }}
                  className="w-7 h-7 flex items-center justify-center rounded hover:bg-[#a0a0b8]/10 transition-colors text-sm"
                >
                  {emoji}
                </button>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  );
}

interface MessageItemProps {
  message: Message;
  onReactionToggle: (
    messageId: number,
    emoji: string,
    reacted: boolean,
  ) => void;
}

export default function MessageItem({
  message,
  onReactionToggle,
}: MessageItemProps) {
  const isAgent = message.author_type === "agent";

  return (
    <div
      className={`group flex gap-3 px-4 py-2 ${
        isAgent
          ? "border-l-2 border-[#e2b714]/20 bg-[#e2b714]/[0.02]"
          : "border-l-2 border-transparent"
      }`}
    >
      {isAgent ? (
        <AgentAvatar name={message.author_name} />
      ) : (
        <HumanAvatar name={message.author_name} />
      )}

      <div className="min-w-0 flex-1">
        <div className="flex items-baseline gap-2">
          <span
            className={`text-sm font-bold ${
              isAgent ? "text-[#e2b714]" : "text-white"
            }`}
          >
            {message.author_name}
          </span>
          {isAgent && (
            <span className="text-[9px] uppercase tracking-widest text-[#e2b714]/40 border border-[#e2b714]/15 rounded px-1 py-px leading-none">
              bot
            </span>
          )}
          <span className="text-[#a0a0b8]/35 text-xs">
            {formatRelativeTime(message.created_at)}
          </span>
        </div>

        <div className="text-[#c8c8e0] mt-0.5">
          <MarkdownRenderer content={message.content} />
        </div>

        <ReactionPills
          reactions={message.reactions}
          onToggle={(emoji, reacted) =>
            onReactionToggle(message.id, emoji, reacted)
          }
        />
      </div>
    </div>
  );
}

export { formatRelativeTime };
