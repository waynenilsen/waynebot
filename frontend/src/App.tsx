import { useState, useEffect, useCallback, useRef } from "react";
import { useAuth } from "./hooks/useAuth";
import { useChannels } from "./hooks/useChannels";
import { useDMs } from "./hooks/useDMs";
import { useMessages } from "./hooks/useMessages";
import { useWebSocket } from "./hooks/useWebSocket";
import { useMentionTargets } from "./hooks/useMentionTargets";
import LoginPage from "./pages/LoginPage";
import PersonaPage from "./pages/PersonaPage";
import AgentDashboard from "./pages/AgentDashboard";
import InvitePage from "./pages/InvitePage";
import ProjectsPage from "./pages/ProjectsPage";
import ProjectsPanel from "./components/ProjectsPanel";
import Layout from "./components/Layout";
import Sidebar from "./components/Sidebar";
import ChannelList from "./components/ChannelList";
import DMList from "./components/DMList";
import { dmDisplayName } from "./components/DMList";
import MessageThread from "./components/MessageThread";
import MessageCompose from "./components/MessageCompose";
import ChannelSwitcher from "./components/ChannelSwitcher";
import MembersPanel from "./components/MembersPanel";

function AuthenticatedApp({
  user,
  logout,
}: {
  user: { id: number; username: string; created_at: string };
  logout: () => Promise<void>;
}) {
  const {
    channels,
    currentChannel,
    currentChannelId,
    selectChannel,
    createChannel,
  } = useChannels();
  const { dms, currentDM, selectDM, createDM } = useDMs();
  const { messages, loading, hasMore, loadMore, sendMessage, toggleReaction } =
    useMessages(currentChannelId);
  const { connected, wasConnected } = useWebSocket(true);
  const { targets: mentionTargets } = useMentionTargets();
  const [currentView, setCurrentView] = useState("channels");
  const [showSwitcher, setShowSwitcher] = useState(false);
  const [showMembers, setShowMembers] = useState(false);
  const [showProjects, setShowProjects] = useState(false);
  const [showConnectedFlash, setShowConnectedFlash] = useState(false);
  const composeRef = useRef<HTMLTextAreaElement>(null);

  // Page title updates
  useEffect(() => {
    if (currentView === "channels" && currentDM) {
      document.title = `${dmDisplayName(currentDM)} - waynebot`;
    } else if (currentView === "channels" && currentChannel) {
      document.title = `# ${currentChannel.name} - waynebot`;
    } else if (currentView === "personas") {
      document.title = "Personas - waynebot";
    } else if (currentView === "agents") {
      document.title = "Agents - waynebot";
    } else if (currentView === "projects") {
      document.title = "Projects - waynebot";
    } else if (currentView === "invites") {
      document.title = "Invites - waynebot";
    } else {
      document.title = "waynebot";
    }
  }, [currentView, currentChannel, currentDM]);

  // Show "connected" flash when reconnection succeeds
  useEffect(() => {
    if (connected && wasConnected) {
      setShowConnectedFlash(true);
      const t = setTimeout(() => setShowConnectedFlash(false), 2000);
      return () => clearTimeout(t);
    }
  }, [connected, wasConnected]);

  // Close side panels when switching channels
  useEffect(() => {
    setShowMembers(false);
    setShowProjects(false);
  }, [currentChannelId]);

  // Focus compose box when switching channels
  useEffect(() => {
    if (currentChannelId && currentView === "channels") {
      // Small delay to let MessageThread render first
      const t = setTimeout(() => composeRef.current?.focus(), 50);
      return () => clearTimeout(t);
    }
  }, [currentChannelId, currentView]);

  // Keyboard shortcuts
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    // Cmd/Ctrl + K → channel switcher
    if ((e.metaKey || e.ctrlKey) && e.key === "k") {
      e.preventDefault();
      setShowSwitcher((prev) => !prev);
    }
  }, []);

  useEffect(() => {
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [handleKeyDown]);

  return (
    <>
      {showSwitcher && (
        <ChannelSwitcher
          channels={channels}
          onSelect={(id) => {
            selectChannel(id);
            setCurrentView("channels");
          }}
          onClose={() => setShowSwitcher(false)}
        />
      )}
      <Layout
        sidebar={
          <Sidebar
            user={user}
            onLogout={logout}
            currentView={currentView}
            onNavigate={setCurrentView}
            channelList={
              <ChannelList
                channels={channels}
                currentChannelId={currentChannelId}
                onSelect={(id) => {
                  selectChannel(id);
                  setCurrentView("channels");
                }}
                onCreate={async (name, desc) => {
                  await createChannel(name, desc);
                }}
              />
            }
            dmList={
              <DMList
                dms={dms}
                currentChannelId={currentChannelId}
                onSelect={(id) => {
                  selectDM(id);
                  setCurrentView("channels");
                }}
                onCreate={async (opts) => {
                  await createDM(opts);
                  setCurrentView("channels");
                }}
              />
            }
          />
        }
      >
        <div className="flex-1 flex flex-col min-h-0">
          {/* WebSocket reconnection indicator */}
          {!connected && (
            <div className="bg-yellow-900/30 border-b border-yellow-600/20 px-4 py-1.5 text-yellow-400 text-xs font-mono flex items-center gap-2">
              <span className="inline-block w-1.5 h-1.5 rounded-full bg-yellow-400 animate-pulse" />
              reconnecting...
            </div>
          )}
          {showConnectedFlash && (
            <div className="bg-emerald-900/30 border-b border-emerald-600/20 px-4 py-1.5 text-emerald-400 text-xs font-mono flex items-center gap-2">
              <span className="inline-block w-1.5 h-1.5 rounded-full bg-emerald-400" />
              connected
            </div>
          )}

          {currentView === "channels" && currentDM ? (
            <>
              <MessageThread
                messages={messages}
                loading={loading}
                hasMore={hasMore}
                onLoadMore={loadMore}
                channelName={dmDisplayName(currentDM)}
                isDM
                onReactionToggle={toggleReaction}
              />
              <MessageCompose
                onSend={sendMessage}
                composeRef={composeRef}
                mentionTargets={mentionTargets}
              />
            </>
          ) : currentView === "channels" && currentChannel ? (
            <div className="flex-1 flex min-h-0">
              <div className="flex-1 flex flex-col min-h-0">
                <MessageThread
                  messages={messages}
                  loading={loading}
                  hasMore={hasMore}
                  onLoadMore={loadMore}
                  channelName={currentChannel.name}
                  onReactionToggle={toggleReaction}
                  onToggleMembers={() => setShowMembers((p) => !p)}
                  onToggleProjects={() => setShowProjects((p) => !p)}
                />
                <MessageCompose
                  onSend={sendMessage}
                  composeRef={composeRef}
                  mentionTargets={mentionTargets}
                />
              </div>
              {showMembers && currentChannelId && (
                <MembersPanel
                  channelId={currentChannelId}
                  onClose={() => setShowMembers(false)}
                />
              )}
              {showProjects && currentChannelId && (
                <ProjectsPanel
                  channelId={currentChannelId}
                  onClose={() => setShowProjects(false)}
                />
              )}
            </div>
          ) : currentView === "channels" ? (
            <div className="flex-1 flex items-center justify-center">
              <div className="text-center">
                <div className="text-[#e2b714]/15 text-4xl mb-3">{">"}_</div>
                <p className="text-[#a0a0b8]/50 text-sm font-mono">
                  select a channel to start chatting
                </p>
                {channels.length > 0 && (
                  <p className="text-[#a0a0b8]/30 text-xs font-mono mt-2">
                    or press{" "}
                    <kbd className="bg-[#0f3460]/50 px-1.5 py-0.5 rounded text-[#a0a0b8]/50">
                      {navigator.platform.includes("Mac") ? "Cmd" : "Ctrl"}+K
                    </kbd>{" "}
                    to search
                  </p>
                )}
              </div>
            </div>
          ) : currentView === "personas" ? (
            <PersonaPage />
          ) : currentView === "agents" ? (
            <AgentDashboard />
          ) : currentView === "projects" ? (
            <ProjectsPage />
          ) : currentView === "invites" ? (
            <InvitePage />
          ) : null}
        </div>
      </Layout>
    </>
  );
}

function App() {
  const { user, loading, login, register, logout } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen bg-[#1a1a2e] flex items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <div className="text-[#e2b714]/30 text-2xl font-mono">{">"}_</div>
          <div className="text-[#e2b714] font-mono text-sm animate-pulse">
            loading...
          </div>
        </div>
      </div>
    );
  }

  if (!user) {
    return <LoginPage onLogin={login} onRegister={register} />;
  }

  return <AuthenticatedApp user={user} logout={logout} />;
}

export default App;
