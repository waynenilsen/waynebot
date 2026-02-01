import type { ReactNode } from "react";
import type { User } from "../types";

interface SidebarProps {
  user: User;
  onLogout: () => void;
  channelList: ReactNode;
  dmList: ReactNode;
  currentView: string;
  onNavigate: (view: string) => void;
}

const navItems = [
  { key: "personas", label: "Personas", icon: "\u25C6" },
  { key: "agents", label: "Agents", icon: "\u25B8" },
  { key: "invites", label: "Invites", icon: "\u2709" },
];

export default function Sidebar({
  user,
  onLogout,
  channelList,
  dmList,
  currentView,
  onNavigate,
}: SidebarProps) {
  return (
    <div className="flex flex-col h-full select-none">
      {/* Brand */}
      <div className="px-4 pt-4 pb-3 border-b border-[#e2b714]/10">
        <h1 className="text-[#e2b714] text-lg font-bold tracking-tight">
          <span className="text-[#e2b714]/40">{">"}</span> waynebot
        </h1>
        <p className="text-[#a0a0b8]/50 text-[10px] tracking-widest uppercase mt-0.5">
          command center
        </p>
      </div>

      {/* Channels & DMs */}
      <div className="flex-1 overflow-y-auto min-h-0">
        <div className="py-2">{channelList}</div>
        <div className="py-2 border-t border-[#e2b714]/10">{dmList}</div>

        {/* Nav section */}
        <div className="px-3 pt-2 pb-2 border-t border-[#e2b714]/10">
          <p className="text-[#a0a0b8]/40 text-[10px] uppercase tracking-widest px-1 mb-1">
            Admin
          </p>
          {navItems.map((item) => {
            const active = currentView === item.key;
            return (
              <button
                key={item.key}
                onClick={() => onNavigate(item.key)}
                className={`w-full text-left px-2 py-1.5 rounded text-sm flex items-center gap-2 transition-colors cursor-pointer ${
                  active
                    ? "text-[#e2b714] bg-[#e2b714]/8"
                    : "text-[#a0a0b8]/70 hover:text-[#a0a0b8] hover:bg-white/3"
                }`}
              >
                <span className="text-xs w-4 text-center opacity-60">
                  {item.icon}
                </span>
                {item.label}
              </button>
            );
          })}
        </div>
      </div>

      {/* User footer */}
      <div className="px-4 py-3 border-t border-[#e2b714]/10 flex items-center justify-between">
        <div className="flex items-center gap-2 min-w-0">
          <div className="w-2 h-2 rounded-full bg-emerald-400 shrink-0" />
          <span className="text-[#a0a0b8] text-sm truncate">
            {user.username}
          </span>
        </div>
        <button
          onClick={onLogout}
          className="text-[#a0a0b8]/40 hover:text-red-400 text-xs transition-colors cursor-pointer"
        >
          logout
        </button>
      </div>
    </div>
  );
}
