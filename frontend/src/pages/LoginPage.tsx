import { useState } from "react";
import type { FormEvent } from "react";
import { getErrorMessage } from "../utils/errors";

interface LoginPageProps {
  onLogin: (username: string, password: string) => Promise<void>;
  onRegister: (
    username: string,
    password: string,
    inviteCode?: string,
  ) => Promise<void>;
}

type Mode = "login" | "register";

export default function LoginPage({ onLogin, onRegister }: LoginPageProps) {
  const [mode, setMode] = useState<Mode>("login");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [inviteCode, setInviteCode] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const usernameValid = /^[a-zA-Z0-9_]{1,50}$/.test(username);
  const passwordValid = password.length >= 8 && password.length <= 128;
  const canSubmit =
    usernameValid && passwordValid && !submitting && username.length > 0;

  function switchMode(next: Mode) {
    setMode(next);
    setError("");
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;
    setError("");
    setSubmitting(true);
    try {
      if (mode === "login") {
        await onLogin(username, password);
      } else {
        await onRegister(username, password, inviteCode || undefined);
      }
    } catch (err: unknown) {
      setError(getErrorMessage(err));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="min-h-screen bg-[#1a1a2e] flex items-center justify-center p-4 relative overflow-hidden">
      {/* Ambient background glow */}
      <div className="absolute inset-0 pointer-events-none">
        <div className="absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] rounded-full bg-[#e2b714]/5 blur-[120px]" />
        <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-[#e2b714]/20 to-transparent" />
      </div>

      <div className="w-full max-w-sm relative z-10">
        {/* Brand */}
        <div className="text-center mb-8">
          <h1 className="text-[#e2b714] font-mono text-3xl font-bold tracking-tight">
            <span className="text-[#e2b714]/40">{">"}</span> waynebot
          </h1>
          <p className="text-[#a0a0b8] text-sm mt-2 tracking-wide">
            {mode === "login"
              ? "Sign in to your workspace"
              : "Create your account"}
          </p>
        </div>

        {/* Card */}
        <div className="bg-[#16213e] border border-[#e2b714]/10 rounded-lg shadow-2xl shadow-black/40">
          {/* Mode tabs */}
          <div className="flex border-b border-[#e2b714]/10">
            <button
              type="button"
              data-testid="tab-login"
              onClick={() => switchMode("login")}
              className={`flex-1 py-3 text-sm font-medium tracking-wide transition-colors relative ${
                mode === "login"
                  ? "text-[#e2b714]"
                  : "text-[#a0a0b8] hover:text-white"
              }`}
            >
              Sign In
              {mode === "login" && (
                <span className="absolute bottom-0 left-4 right-4 h-0.5 bg-[#e2b714] rounded-full" />
              )}
            </button>
            <button
              type="button"
              data-testid="tab-register"
              onClick={() => switchMode("register")}
              className={`flex-1 py-3 text-sm font-medium tracking-wide transition-colors relative ${
                mode === "register"
                  ? "text-[#e2b714]"
                  : "text-[#a0a0b8] hover:text-white"
              }`}
            >
              Register
              {mode === "register" && (
                <span className="absolute bottom-0 left-4 right-4 h-0.5 bg-[#e2b714] rounded-full" />
              )}
            </button>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="p-6 space-y-4">
            {/* Error */}
            {error && (
              <div
                role="alert"
                className="bg-red-500/10 border border-red-500/20 text-red-400 text-sm px-3 py-2 rounded"
              >
                {error}
              </div>
            )}

            {/* Username */}
            <div>
              <label
                htmlFor="username"
                className="block text-[#a0a0b8] text-xs font-medium uppercase tracking-wider mb-1.5"
              >
                Username
              </label>
              <input
                id="username"
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="your_username"
                autoComplete="username"
                maxLength={50}
                className="w-full bg-[#0f3460]/50 border border-[#e2b714]/10 rounded px-3 py-2.5 text-white text-sm placeholder-[#a0a0b8]/40 focus:outline-none focus:border-[#e2b714]/40 focus:ring-1 focus:ring-[#e2b714]/20 transition-colors font-mono"
              />
              {username.length > 0 && !usernameValid && (
                <p className="text-red-400/80 text-xs mt-1">
                  Letters, numbers, and underscores only (1-50 chars)
                </p>
              )}
            </div>

            {/* Password */}
            <div>
              <label
                htmlFor="password"
                className="block text-[#a0a0b8] text-xs font-medium uppercase tracking-wider mb-1.5"
              >
                Password
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                autoComplete={
                  mode === "login" ? "current-password" : "new-password"
                }
                maxLength={128}
                className="w-full bg-[#0f3460]/50 border border-[#e2b714]/10 rounded px-3 py-2.5 text-white text-sm placeholder-[#a0a0b8]/40 focus:outline-none focus:border-[#e2b714]/40 focus:ring-1 focus:ring-[#e2b714]/20 transition-colors"
              />
              {password.length > 0 && !passwordValid && (
                <p className="text-red-400/80 text-xs mt-1">
                  Must be 8-128 characters
                </p>
              )}
            </div>

            {/* Invite code (register only) */}
            {mode === "register" && (
              <div>
                <label
                  htmlFor="inviteCode"
                  className="block text-[#a0a0b8] text-xs font-medium uppercase tracking-wider mb-1.5"
                >
                  Invite Code{" "}
                  <span className="text-[#a0a0b8]/50 normal-case tracking-normal">
                    (optional)
                  </span>
                </label>
                <input
                  id="inviteCode"
                  type="text"
                  value={inviteCode}
                  onChange={(e) => setInviteCode(e.target.value)}
                  placeholder="abc-123-xyz"
                  autoComplete="off"
                  className="w-full bg-[#0f3460]/50 border border-[#e2b714]/10 rounded px-3 py-2.5 text-white text-sm placeholder-[#a0a0b8]/40 focus:outline-none focus:border-[#e2b714]/40 focus:ring-1 focus:ring-[#e2b714]/20 transition-colors font-mono"
                />
              </div>
            )}

            {/* Submit */}
            <button
              type="submit"
              disabled={!canSubmit}
              className="w-full bg-[#e2b714] hover:bg-[#c9a212] disabled:bg-[#e2b714]/20 disabled:text-[#a0a0b8]/40 text-[#1a1a2e] font-semibold text-sm py-2.5 rounded transition-colors mt-2 cursor-pointer disabled:cursor-not-allowed"
            >
              {submitting ? (
                <span className="inline-flex items-center gap-2">
                  <svg
                    className="animate-spin h-4 w-4"
                    viewBox="0 0 24 24"
                    fill="none"
                  >
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                    />
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                    />
                  </svg>
                  {mode === "login" ? "Signing in..." : "Creating account..."}
                </span>
              ) : mode === "login" ? (
                "Sign In"
              ) : (
                "Create Account"
              )}
            </button>
          </form>
        </div>

        {/* Footer hint */}
        <p className="text-center text-[#a0a0b8]/40 text-xs mt-6">
          {mode === "login" ? (
            <>
              No account?{" "}
              <button
                type="button"
                onClick={() => switchMode("register")}
                className="text-[#e2b714]/60 hover:text-[#e2b714] transition-colors cursor-pointer"
              >
                Register
              </button>
            </>
          ) : (
            <>
              Already have an account?{" "}
              <button
                type="button"
                onClick={() => switchMode("login")}
                className="text-[#e2b714]/60 hover:text-[#e2b714] transition-colors cursor-pointer"
              >
                Sign in
              </button>
            </>
          )}
        </p>
      </div>
    </div>
  );
}
