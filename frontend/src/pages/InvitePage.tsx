import { useState } from "react";
import { useInvites } from "../hooks/useInvites";

export default function InvitePage() {
  const { invites, loading, createInvite } = useInvites();
  const [generating, setGenerating] = useState(false);
  const [copied, setCopied] = useState<string | null>(null);

  async function handleGenerate() {
    setGenerating(true);
    try {
      await createInvite();
    } catch {
      // ignore
    } finally {
      setGenerating(false);
    }
  }

  async function handleCopy(code: string) {
    await navigator.clipboard.writeText(code);
    setCopied(code);
    setTimeout(() => setCopied(null), 2000);
  }

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* Header */}
      <div className="px-6 py-4 border-b border-[#e2b714]/10 flex items-center justify-between shrink-0">
        <div>
          <h1 className="text-white text-lg font-bold font-mono flex items-center gap-2">
            <span className="text-[#e2b714]/40 text-sm">&#9993;</span>
            Invites
          </h1>
          <p className="text-[#a0a0b8]/50 text-xs font-mono mt-0.5">
            {invites.length} invite{invites.length !== 1 && "s"} generated
          </p>
        </div>
        <button
          onClick={handleGenerate}
          disabled={generating}
          className="bg-[#e2b714] hover:bg-[#c9a212] disabled:bg-[#e2b714]/20 disabled:text-[#a0a0b8]/40 text-[#1a1a2e] font-semibold text-sm py-2 px-4 rounded transition-colors cursor-pointer disabled:cursor-not-allowed"
        >
          {generating ? "Generating..." : "+ Generate Invite"}
        </button>
      </div>

      {/* List */}
      <div className="flex-1 overflow-y-auto p-6">
        {loading && invites.length === 0 ? (
          <div className="text-[#a0a0b8]/50 text-sm font-mono text-center py-12">
            loading...
          </div>
        ) : invites.length === 0 ? (
          <div className="text-[#a0a0b8]/50 text-sm font-mono text-center py-12">
            no invites yet — generate one to share
          </div>
        ) : (
          <div className="space-y-2 max-w-2xl mx-auto">
            {invites.map((inv) => {
              const used = inv.used_by !== null;
              return (
                <div
                  key={inv.id}
                  className={`bg-[#16213e] border rounded-lg px-4 py-3 flex items-center justify-between gap-4 ${
                    used
                      ? "border-[#a0a0b8]/10 opacity-60"
                      : "border-[#e2b714]/10"
                  }`}
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <div
                      className={`w-2 h-2 rounded-full shrink-0 ${
                        used ? "bg-[#a0a0b8]/30" : "bg-[#e2b714]/60"
                      }`}
                    />
                    <code className="text-white text-sm font-mono truncate">
                      {inv.code}
                    </code>
                  </div>

                  <div className="flex items-center gap-3 shrink-0">
                    <span
                      className={`text-[10px] uppercase tracking-widest font-mono ${
                        used ? "text-[#a0a0b8]/30" : "text-[#e2b714]/40"
                      }`}
                    >
                      {used ? "used" : "available"}
                    </span>
                    {!used && (
                      <button
                        onClick={() => handleCopy(inv.code)}
                        className="text-[#a0a0b8]/50 hover:text-[#e2b714] text-xs font-mono px-2 py-1 rounded hover:bg-[#e2b714]/5 transition-colors cursor-pointer"
                      >
                        {copied === inv.code ? "copied!" : "copy"}
                      </button>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
