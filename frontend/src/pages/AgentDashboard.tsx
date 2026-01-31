import { useEffect } from "react";
import { useAgents } from "../hooks/useAgents";

export default function AgentDashboard() {
  const { agents, loading, startAgents, stopAgents, refresh } = useAgents();

  useEffect(() => {
    refresh();
  }, [refresh]);

  const running = agents.filter((a) => a.status === "running").length;
  const stopped = agents.length - running;

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      {/* Header */}
      <div className="px-6 py-4 border-b border-[#e2b714]/10 flex items-center justify-between shrink-0">
        <div>
          <h1 className="text-white text-lg font-bold font-mono flex items-center gap-2">
            <span className="text-[#e2b714]/40 text-sm">&#9656;</span>
            Agents
          </h1>
          <p className="text-[#a0a0b8]/50 text-xs font-mono mt-0.5">
            {running} running, {stopped} stopped
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={refresh}
            disabled={loading}
            className="text-[#a0a0b8]/60 hover:text-[#a0a0b8] text-sm py-2 px-3 rounded border border-[#e2b714]/10 hover:border-[#e2b714]/25 transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed font-mono"
          >
            {loading ? "refreshing..." : "refresh"}
          </button>
          <button
            onClick={startAgents}
            className="bg-emerald-500/20 hover:bg-emerald-500/30 text-emerald-400 font-semibold text-sm py-2 px-4 rounded transition-colors cursor-pointer border border-emerald-500/20 font-mono"
          >
            start all
          </button>
          <button
            onClick={stopAgents}
            className="bg-red-500/10 hover:bg-red-500/20 text-red-400 font-semibold text-sm py-2 px-4 rounded transition-colors cursor-pointer border border-red-500/20 font-mono"
          >
            stop all
          </button>
        </div>
      </div>

      {/* Agent list */}
      <div className="flex-1 overflow-y-auto p-6">
        {loading && agents.length === 0 ? (
          <div className="text-[#a0a0b8]/50 text-sm font-mono text-center py-12">
            loading...
          </div>
        ) : agents.length === 0 ? (
          <div className="text-[#a0a0b8]/50 text-sm font-mono text-center py-12">
            no agents configured — create a persona first
          </div>
        ) : (
          <div className="grid gap-3 max-w-3xl mx-auto">
            {agents.map((agent) => {
              const isRunning = agent.status === "running";
              return (
                <div
                  key={agent.persona_id}
                  className={`bg-[#16213e] border rounded-lg p-4 ${
                    isRunning ? "border-emerald-500/20" : "border-[#e2b714]/10"
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3 min-w-0">
                      <div
                        className={`w-2.5 h-2.5 rounded-full shrink-0 ${
                          isRunning
                            ? "bg-emerald-400 shadow-[0_0_6px_rgba(52,211,153,0.4)]"
                            : "bg-[#a0a0b8]/30"
                        }`}
                      />
                      <div className="min-w-0">
                        <h3 className="text-white font-bold font-mono text-sm truncate">
                          {agent.persona_name}
                        </h3>
                        <span
                          className={`text-[10px] uppercase tracking-widest font-mono ${
                            isRunning
                              ? "text-emerald-400/60"
                              : "text-[#a0a0b8]/30"
                          }`}
                        >
                          {agent.status}
                        </span>
                      </div>
                    </div>

                    {agent.channels.length > 0 && (
                      <div className="flex items-center gap-1.5 shrink-0">
                        {agent.channels.map((ch) => (
                          <span
                            key={ch}
                            className="text-[#a0a0b8]/40 text-xs font-mono bg-[#0f3460]/50 px-2 py-0.5 rounded"
                          >
                            #{ch}
                          </span>
                        ))}
                      </div>
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
