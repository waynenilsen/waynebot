import {
  useState,
  useRef,
  useCallback,
  useEffect,
  type KeyboardEvent,
  type RefObject,
} from "react";
import type { MentionTarget } from "../types";

const MAX_LENGTH = 10_000;

interface MessageComposeProps {
  onSend: (content: string) => Promise<void>;
  composeRef?: RefObject<HTMLTextAreaElement | null>;
  mentionTargets?: MentionTarget[];
}

interface MentionState {
  active: boolean;
  query: string;
  startPos: number; // position of the '@' in text
  selectedIndex: number;
}

function filterTargets(
  targets: MentionTarget[],
  query: string,
): MentionTarget[] {
  if (!query) return targets.slice(0, 8);
  const lower = query.toLowerCase();
  return targets
    .filter((t) => t.name.toLowerCase().includes(lower))
    .slice(0, 8);
}

export default function MessageCompose({
  onSend,
  composeRef,
  mentionTargets = [],
}: MessageComposeProps) {
  const [text, setText] = useState("");
  const [sending, setSending] = useState(false);
  const [mention, setMention] = useState<MentionState>({
    active: false,
    query: "",
    startPos: 0,
    selectedIndex: 0,
  });
  const internalRef = useRef<HTMLTextAreaElement>(null);
  const textareaRef = composeRef ?? internalRef;
  const dropdownRef = useRef<HTMLDivElement>(null);

  const trimmed = text.trim();
  const canSend = trimmed.length > 0 && !sending;
  const charCount = text.length;
  const nearLimit = charCount > MAX_LENGTH * 0.9;
  const overLimit = charCount > MAX_LENGTH;

  const filtered = mention.active
    ? filterTargets(mentionTargets, mention.query)
    : [];

  const autoResize = useCallback(() => {
    const el = textareaRef.current;
    if (!el) return;
    el.style.height = "auto";
    el.style.height = Math.min(el.scrollHeight, 200) + "px";
  }, [textareaRef]);

  const completeMention = useCallback(
    (target: MentionTarget) => {
      const before = text.slice(0, mention.startPos);
      const after = text.slice(mention.startPos + 1 + mention.query.length);
      const newText = before + "@" + target.name + " " + after;
      setText(newText);
      setMention({ active: false, query: "", startPos: 0, selectedIndex: 0 });

      // Restore cursor position after the inserted mention.
      const cursorPos = mention.startPos + 1 + target.name.length + 1;
      requestAnimationFrame(() => {
        const el = textareaRef.current;
        if (el) {
          el.selectionStart = cursorPos;
          el.selectionEnd = cursorPos;
          el.focus();
        }
      });
    },
    [text, mention, textareaRef],
  );

  const handleSend = useCallback(async () => {
    if (!canSend || overLimit) return;
    setSending(true);
    try {
      await onSend(trimmed);
      setText("");
      setMention({ active: false, query: "", startPos: 0, selectedIndex: 0 });
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
      if (mention.active && filtered.length > 0) {
        if (e.key === "ArrowDown") {
          e.preventDefault();
          setMention((m) => ({
            ...m,
            selectedIndex: (m.selectedIndex + 1) % filtered.length,
          }));
          return;
        }
        if (e.key === "ArrowUp") {
          e.preventDefault();
          setMention((m) => ({
            ...m,
            selectedIndex:
              (m.selectedIndex - 1 + filtered.length) % filtered.length,
          }));
          return;
        }
        if (e.key === "Tab" || e.key === "Enter") {
          e.preventDefault();
          completeMention(filtered[mention.selectedIndex]);
          return;
        }
        if (e.key === "Escape") {
          e.preventDefault();
          setMention({
            active: false,
            query: "",
            startPos: 0,
            selectedIndex: 0,
          });
          return;
        }
      }

      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        handleSend();
      }
    },
    [handleSend, mention, filtered, completeMention],
  );

  const handleChange = useCallback(
    (value: string) => {
      setText(value);
      autoResize();

      const el = textareaRef.current;
      if (!el) return;

      const cursorPos = el.selectionStart;
      // Look backwards from cursor for an '@' that starts a mention.
      const textBeforeCursor = value.slice(0, cursorPos);
      const atIndex = textBeforeCursor.lastIndexOf("@");

      if (atIndex === -1) {
        if (mention.active) {
          setMention({
            active: false,
            query: "",
            startPos: 0,
            selectedIndex: 0,
          });
        }
        return;
      }

      // '@' must be at start of text or preceded by whitespace.
      const charBefore = atIndex > 0 ? value[atIndex - 1] : " ";
      if (!/\s/.test(charBefore)) {
        if (mention.active) {
          setMention({
            active: false,
            query: "",
            startPos: 0,
            selectedIndex: 0,
          });
        }
        return;
      }

      const query = textBeforeCursor.slice(atIndex + 1);
      // No spaces in mention query.
      if (/\s/.test(query)) {
        if (mention.active) {
          setMention({
            active: false,
            query: "",
            startPos: 0,
            selectedIndex: 0,
          });
        }
        return;
      }

      setMention({ active: true, query, startPos: atIndex, selectedIndex: 0 });
    },
    [textareaRef, autoResize, mention.active],
  );

  // Scroll selected item into view.
  useEffect(() => {
    if (!mention.active || !dropdownRef.current) return;
    const selected = dropdownRef.current.children[mention.selectedIndex] as
      | HTMLElement
      | undefined;
    selected?.scrollIntoView({ block: "nearest" });
  }, [mention.active, mention.selectedIndex]);

  return (
    <div className="shrink-0 border-t border-[#e2b714]/10 bg-[#1a1a2e] px-4 py-3">
      <div className="relative">
        {/* Mention autocomplete dropdown */}
        {mention.active && filtered.length > 0 && (
          <div
            ref={dropdownRef}
            className="absolute bottom-full left-0 right-0 mb-1 bg-[#16213e] border border-[#a0a0b8]/15 rounded-lg shadow-lg overflow-hidden max-h-48 overflow-y-auto z-50"
          >
            {filtered.map((target, i) => (
              <button
                key={`${target.type}-${target.id}`}
                onMouseDown={(e) => {
                  e.preventDefault(); // prevent textarea blur
                  completeMention(target);
                }}
                className={`w-full flex items-center gap-2 px-3 py-1.5 text-left text-sm font-mono transition-colors ${
                  i === mention.selectedIndex
                    ? "bg-[#e2b714]/15 text-[#e2b714]"
                    : "text-[#c8c8e0] hover:bg-[#a0a0b8]/10"
                }`}
              >
                {target.type === "persona" ? (
                  <span className="w-4 h-4 shrink-0 flex items-center justify-center">
                    <span className="inline-block w-3 h-3 rotate-45 bg-[#e2b714]/15 border border-[#e2b714]/30" />
                  </span>
                ) : (
                  <span className="w-4 h-4 rounded-full bg-[#0f3460] border border-[#a0a0b8]/15 flex items-center justify-center shrink-0">
                    <span className="text-[#a0a0b8] text-[8px] font-bold uppercase">
                      {target.name.charAt(0)}
                    </span>
                  </span>
                )}
                <span>@{target.name}</span>
                <span className="ml-auto text-[10px] text-[#a0a0b8]/40">
                  {target.type}
                </span>
              </button>
            ))}
          </div>
        )}

        <div className="flex gap-2 items-end">
          <textarea
            ref={textareaRef}
            value={text}
            onChange={(e) => handleChange(e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={sending}
            placeholder="type a message... (@mention to ping)"
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
