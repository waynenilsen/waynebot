import { type ReactNode } from "react";
import Markdown from "react-markdown";

interface MarkdownRendererProps {
  content: string;
}

const MENTION_RE = /(^|\s)(@[a-zA-Z0-9_]+)/g;

/** Walks children and highlights @mentions within text nodes. */
function highlightMentions(children: ReactNode): ReactNode {
  if (typeof children === "string") {
    const parts: ReactNode[] = [];
    let lastIndex = 0;
    let match: RegExpExecArray | null;
    MENTION_RE.lastIndex = 0;

    while ((match = MENTION_RE.exec(children)) !== null) {
      const prefix = match[1]; // whitespace or empty before the @
      const mention = match[2]; // @username
      const start = match.index + prefix.length;

      if (start > lastIndex) {
        parts.push(children.slice(lastIndex, start));
      }
      parts.push(
        <span
          key={start}
          className="text-[#e2b714] font-bold bg-[#e2b714]/10 rounded px-0.5"
        >
          {mention}
        </span>,
      );
      lastIndex = start + mention.length;
    }

    if (parts.length === 0) return children;
    if (lastIndex < children.length) {
      parts.push(children.slice(lastIndex));
    }
    return <>{parts}</>;
  }

  if (Array.isArray(children)) {
    return children.map((child, i) => {
      if (typeof child === "string") {
        return <span key={i}>{highlightMentions(child)}</span>;
      }
      return child;
    });
  }

  return children;
}

export default function MarkdownRenderer({ content }: MarkdownRendererProps) {
  return (
    <div className="markdown-content text-sm leading-relaxed">
      <Markdown
        components={{
          a: ({ href, children }) => (
            <a
              href={href}
              target="_blank"
              rel="noopener noreferrer"
              className="text-[#e2b714] hover:underline"
            >
              {children}
            </a>
          ),
          code: ({ className, children }) => {
            const isBlock = className?.startsWith("language-");
            if (isBlock) {
              return (
                <code className="block bg-[#0f3460]/60 rounded px-3 py-2 my-2 text-xs overflow-x-auto border border-[#e2b714]/10 text-[#c8c8e0]">
                  {children}
                </code>
              );
            }
            return (
              <code className="bg-[#0f3460]/60 rounded px-1.5 py-0.5 text-xs text-[#c8c8e0]">
                {children}
              </code>
            );
          },
          pre: ({ children }) => <pre className="my-1">{children}</pre>,
          strong: ({ children }) => (
            <strong className="font-bold text-white">{children}</strong>
          ),
          em: ({ children }) => (
            <em className="italic text-[#a0a0b8]">{children}</em>
          ),
          ul: ({ children }) => (
            <ul className="list-disc list-inside my-1 space-y-0.5">
              {children}
            </ul>
          ),
          ol: ({ children }) => (
            <ol className="list-decimal list-inside my-1 space-y-0.5">
              {children}
            </ol>
          ),
          p: ({ children }) => (
            <p className="my-1 first:mt-0 last:mb-0">
              {highlightMentions(children)}
            </p>
          ),
          li: ({ children }) => <li>{highlightMentions(children)}</li>,
        }}
      >
        {content}
      </Markdown>
    </div>
  );
}
