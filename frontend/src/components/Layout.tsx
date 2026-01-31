import type { ReactNode } from "react";

interface LayoutProps {
  sidebar: ReactNode;
  children: ReactNode;
}

export default function Layout({ sidebar, children }: LayoutProps) {
  return (
    <div className="h-screen flex bg-[#1a1a2e] overflow-hidden font-mono">
      {/* Sidebar */}
      <div className="w-60 shrink-0 flex flex-col bg-[#131a2e] border-r border-[#e2b714]/15">
        {sidebar}
      </div>

      {/* Main content */}
      <div className="flex-1 flex flex-col min-w-0">{children}</div>
    </div>
  );
}
