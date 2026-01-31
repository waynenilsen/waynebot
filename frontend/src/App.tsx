import { useAuth } from "./hooks/useAuth";
import LoginPage from "./pages/LoginPage";

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

  return (
    <div className="min-h-screen bg-[#1a1a2e] flex items-center justify-center">
      <div className="text-center">
        <h1 className="text-white text-2xl font-mono">
          Welcome, <span className="text-[#e2b714]">{user.username}</span>
        </h1>
        <button
          onClick={logout}
          className="mt-4 text-[#a0a0b8] hover:text-white text-sm transition-colors cursor-pointer"
        >
          Sign out
        </button>
      </div>
    </div>
  );
}

export default App;
