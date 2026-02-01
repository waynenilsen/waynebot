interface TypingAgent {
  persona_id: number;
  persona_name: string;
}

interface TypingIndicatorProps {
  agents: TypingAgent[];
}

export default function TypingIndicator({ agents }: TypingIndicatorProps) {
  if (agents.length === 0) return null;

  const names =
    agents.length === 1
      ? agents[0].persona_name
      : agents.map((a) => a.persona_name).join(", ");

  const verb = agents.length === 1 ? "is" : "are";

  return (
    <div className="flex items-center gap-2 px-4 py-1.5 text-xs font-mono text-[#e2b714]/60">
      <span className="flex gap-0.5">
        <span className="w-1.5 h-1.5 rounded-full bg-[#e2b714]/50 animate-bounce [animation-delay:0ms]" />
        <span className="w-1.5 h-1.5 rounded-full bg-[#e2b714]/50 animate-bounce [animation-delay:150ms]" />
        <span className="w-1.5 h-1.5 rounded-full bg-[#e2b714]/50 animate-bounce [animation-delay:300ms]" />
      </span>
      <span>
        {names} {verb} typing...
      </span>
    </div>
  );
}
