import { type ReactNode, useState } from "react";
import Markdown from "react-markdown";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { oneDark } from "react-syntax-highlighter/dist/esm/styles/prism";

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

const codeBlockStyle: Record<string, React.CSSProperties> = {
  ...oneDark,
  'pre[class*="language-"]': {
    ...oneDark['pre[class*="language-"]'],
    background: "rgba(15, 52, 96, 0.6)",
    margin: 0,
    padding: "0.75rem 1rem",
    fontSize: "0.75rem",
    lineHeight: "1.5",
    borderRadius: 0,
  },
  'code[class*="language-"]': {
    ...oneDark['code[class*="language-"]'],
    background: "none",
    fontSize: "0.75rem",
  },
};

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  return (
    <button
      onClick={handleCopy}
      className="text-[10px] px-1.5 py-0.5 rounded bg-[#0f3460]/80 text-[#a0a0b8] hover:text-[#e2b714] hover:bg-[#0f3460] transition-colors cursor-pointer"
    >
      {copied ? "Copied!" : "Copy"}
    </button>
  );
}

function extractText(children: ReactNode): string {
  if (typeof children === "string") return children;
  if (Array.isArray(children)) return children.map(extractText).join("");
  if (children && typeof children === "object" && "props" in children) {
    const props = (children as { props: { children?: ReactNode } }).props;
    if (props?.children != null) return extractText(props.children);
  }
  return "";
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
            const langMatch = className?.match(/language-(\w+)/);
            if (langMatch) {
              const language = langMatch[1];
              const codeString = extractText(children).replace(/\n$/, "");
              return (
                <div className="rounded overflow-hidden border border-[#e2b714]/10 my-2">
                  <div className="flex items-center justify-between px-3 py-1 bg-[#0f3460]/80 border-b border-[#e2b714]/10">
                    <span className="text-[10px] font-mono text-[#a0a0b8]/70 uppercase tracking-wider">
                      {language}
                    </span>
                    <CopyButton text={codeString} />
                  </div>
                  <SyntaxHighlighter
                    style={codeBlockStyle}
                    language={language}
                    PreTag="div"
                  >
                    {codeString}
                  </SyntaxHighlighter>
                </div>
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
