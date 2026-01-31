import {
  useState,
  useRef,
  useCallback,
  type KeyboardEvent,
  type RefObject,
} from "react";

const MAX_LENGTH = 10_000;

interface MessageComposeProps {
  onSend: (content: string) => Promise<void>;
  composeRef?: RefObject<HTMLTextAreaElement | null>;
}

export default function MessageCompose({
  onSend,
  composeRef,
}: MessageComposeProps) {
  const [text, setText] = useState("");
  const [sending, setSending] = useState(false);
  const internalRef = useRef<HTMLTextAreaElement>(null);
  const textareaRef = composeRef ?? internalRef;

  const trimmed = text.trim();
  const canSend = trimmed.length > 0 && !sending;
  const charCount = text.length;
  const nearLimit = charCount > MAX_LENGTH * 0.9;
  const overLimit = charCount > MAX_LENGTH;

  const autoResize = useCallback(() => {
    const el = textareaRef.current;
    if (!el) return;
    el.style.height = "auto";
    el.style.height = Math.min(el.scrollHeight, 200) + "px";
  }, [textareaRef]);

  const handleSend = useCallback(async () => {
    if (!canSend || overLimit) return;
    setSending(true);
    try {
      await onSend(trimmed);
      setText("");
      if (textareaRef.current) {
        textareaRef.current.style.height = "auto";
      }
    } finally {
      setSending(false);
      textareaRef.current?.focus();
    }
  }, [canSend, overLimit, onSend, trimmed, textareaRef]);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        handleSend();
      }
    },
    [handleSend],
  );

  return (
    <div className="shrink-0 border-t border-[#e2b714]/10 bg-[#1a1a2e] px-4 py-3">
      <div className="flex gap-2 items-end">
        <textarea
          ref={textareaRef}
          value={text}
          onChange={(e) => {
            setText(e.target.value);
            autoResize();
          }}
          onKeyDown={handleKeyDown}
          disabled={sending}
          placeholder="type a message..."
          rows={1}
          className="flex-1 bg-[#0f3460]/40 border border-[#a0a0b8]/10 rounded px-3 py-2 text-sm font-mono text-[#c8c8e0] placeholder:text-[#a0a0b8]/30 focus:outline-none focus:border-[#e2b714]/30 resize-none disabled:opacity-50 transition-colors"
        />
        <button
          onClick={handleSend}
          disabled={!canSend || overLimit}
          className="shrink-0 px-4 py-2 bg-[#e2b714]/15 border border-[#e2b714]/25 rounded text-[#e2b714] text-sm font-mono font-bold hover:bg-[#e2b714]/25 transition-colors disabled:opacity-30 disabled:cursor-default"
        >
          {sending ? "..." : "send"}
        </button>
      </div>

      {/* Character count — only visible near limit */}
      {nearLimit && (
        <div className="mt-1 text-right">
          <span
            className={`text-xs font-mono ${
              overLimit ? "text-red-400" : "text-[#a0a0b8]/40"
            }`}
          >
            {charCount.toLocaleString()} / {MAX_LENGTH.toLocaleString()}
          </span>
        </div>
      )}
    </div>
  );
}
