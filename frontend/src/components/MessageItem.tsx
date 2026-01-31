import type { Message } from "../types";
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

interface MessageItemProps {
  message: Message;
}

export default function MessageItem({ message }: MessageItemProps) {
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
      </div>
    </div>
  );
}

export { formatRelativeTime };
