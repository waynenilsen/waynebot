import { useState } from "react";
import { useAuth } from "./hooks/useAuth";
import { useChannels } from "./hooks/useChannels";
import { useMessages } from "./hooks/useMessages";
import { useWebSocket } from "./hooks/useWebSocket";
import LoginPage from "./pages/LoginPage";
import PersonaPage from "./pages/PersonaPage";
import AgentDashboard from "./pages/AgentDashboard";
import InvitePage from "./pages/InvitePage";
import Layout from "./components/Layout";
import Sidebar from "./components/Sidebar";
import ChannelList from "./components/ChannelList";
import MessageThread from "./components/MessageThread";
import MessageCompose from "./components/MessageCompose";

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
  const { messages, loading, hasMore, loadMore, sendMessage } =
    useMessages(currentChannelId);
  const { connected } = useWebSocket(true);
  const [currentView, setCurrentView] = useState("channels");

  return (
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
        />
      }
    >
      <div className="flex-1 flex flex-col">
        {!connected && (
          <div className="bg-yellow-900/30 border-b border-yellow-600/20 px-4 py-1.5 text-yellow-400 text-xs font-mono">
            reconnecting...
          </div>
        )}

        {currentView === "channels" && currentChannel ? (
          <>
            <MessageThread
              messages={messages}
              loading={loading}
              hasMore={hasMore}
              onLoadMore={loadMore}
              channelName={currentChannel.name}
            />
            <MessageCompose onSend={sendMessage} />
          </>
        ) : currentView === "channels" ? (
          <div className="flex-1 flex items-center justify-center">
            <div className="text-[#a0a0b8]/50 text-sm font-mono">
              select a channel to start chatting
            </div>
          </div>
        ) : currentView === "personas" ? (
          <PersonaPage />
        ) : currentView === "agents" ? (
          <AgentDashboard />
        ) : currentView === "invites" ? (
          <InvitePage />
        ) : null}
      </div>
    </Layout>
  );
}

function App() {
  const { user, loading, login, register, logout } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen bg-[#1a1a2e] flex items-center justify-center">
        <div className="text-[#e2b714] font-mono text-lg animate-pulse">
          loading...
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
