import { useState } from "react";
import { useAuth } from "./hooks/useAuth";
import { useChannels } from "./hooks/useChannels";
import { useWebSocket } from "./hooks/useWebSocket";
import LoginPage from "./pages/LoginPage";
import Layout from "./components/Layout";
import Sidebar from "./components/Sidebar";
import ChannelList from "./components/ChannelList";

function AuthenticatedApp({
  user,
  logout,
}: {
  user: { id: number; username: string; created_at: string };
  logout: () => Promise<void>;
}) {
  const { channels, currentChannelId, selectChannel, createChannel } =
    useChannels();
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

        <div className="flex-1 flex items-center justify-center">
          {currentView === "channels" && currentChannelId ? (
            <div className="text-[#a0a0b8]/50 text-sm font-mono">
              # channel view coming soon
            </div>
          ) : currentView === "channels" ? (
            <div className="text-[#a0a0b8]/50 text-sm font-mono">
              select a channel to start chatting
            </div>
          ) : (
            <div className="text-[#a0a0b8]/50 text-sm font-mono">
              {currentView} page coming soon
            </div>
          )}
        </div>
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
