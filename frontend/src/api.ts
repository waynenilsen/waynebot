import type {
  AgentStatus,
  AuthResponse,
  Channel,
  Invite,
  Message,
  Persona,
  User,
} from "./types";
import { clearToken, getToken, setToken } from "./utils/token";

class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function apiFetch<T>(path: string, opts?: RequestInit): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(opts?.headers as Record<string, string>),
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(path, { ...opts, headers });
  if (!res.ok) {
    let msg = res.statusText;
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) msg = body.error;
    } catch {
      // ignore parse errors
    }
    throw new ApiError(msg, res.status);
  }

  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

export async function register(
  username: string,
  password: string,
  inviteCode?: string,
): Promise<AuthResponse> {
  const resp = await apiFetch<AuthResponse>("/api/auth/register", {
    method: "POST",
    body: JSON.stringify({
      username,
      password,
      invite_code: inviteCode ?? "",
    }),
  });
  setToken(resp.token);
  return resp;
}

export async function login(
  username: string,
  password: string,
): Promise<AuthResponse> {
  const resp = await apiFetch<AuthResponse>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ username, password }),
  });
  setToken(resp.token);
  return resp;
}

export async function logout(): Promise<void> {
  await apiFetch<void>("/api/auth/logout", { method: "POST" });
  clearToken();
}

export async function getMe(): Promise<User> {
  return apiFetch<User>("/api/auth/me");
}

export async function getChannels(): Promise<Channel[]> {
  return apiFetch<Channel[]>("/api/channels");
}

export async function createChannel(
  name: string,
  description: string,
): Promise<Channel> {
  return apiFetch<Channel>("/api/channels", {
    method: "POST",
    body: JSON.stringify({ name, description }),
  });
}

export async function getMessages(
  channelId: number,
  opts?: { limit?: number; before?: number },
): Promise<Message[]> {
  const params = new URLSearchParams();
  if (opts?.limit) params.set("limit", String(opts.limit));
  if (opts?.before) params.set("before", String(opts.before));
  const qs = params.toString();
  return apiFetch<Message[]>(
    `/api/channels/${channelId}/messages${qs ? `?${qs}` : ""}`,
  );
}

export async function postMessage(
  channelId: number,
  content: string,
): Promise<Message> {
  return apiFetch<Message>(`/api/channels/${channelId}/messages`, {
    method: "POST",
    body: JSON.stringify({ content }),
  });
}

export async function getPersonas(): Promise<Persona[]> {
  return apiFetch<Persona[]>("/api/personas");
}

export async function createPersona(
  data: Omit<Persona, "id" | "created_at">,
): Promise<Persona> {
  return apiFetch<Persona>("/api/personas", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updatePersona(
  id: number,
  data: Omit<Persona, "id" | "created_at">,
): Promise<Persona> {
  return apiFetch<Persona>(`/api/personas/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export async function deletePersona(id: number): Promise<void> {
  return apiFetch<void>(`/api/personas/${id}`, { method: "DELETE" });
}

export async function getInvites(): Promise<Invite[]> {
  return apiFetch<Invite[]>("/api/invites");
}

export async function createInvite(): Promise<Invite> {
  return apiFetch<Invite>("/api/invites", { method: "POST" });
}

export async function getAgentStatus(): Promise<AgentStatus[]> {
  return apiFetch<AgentStatus[]>("/api/agents/status");
}

export async function startAgents(): Promise<void> {
  return apiFetch<void>("/api/agents/start", { method: "POST" });
}

export async function stopAgents(): Promise<void> {
  return apiFetch<void>("/api/agents/stop", { method: "POST" });
}

export { ApiError };
