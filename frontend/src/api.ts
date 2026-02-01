import type {
  AgentStatsResponse,
  AgentStatusResponse,
  AuthResponse,
  Channel,
  ChannelMember,
  ContextBudget,
  DMChannel,
  ReactionCount,
  Invite,
  LLMCall,
  MentionTarget,
  Message,
  Persona,
  PersonaTemplate,
  Project,
  ProjectDocument,
  ProjectDocumentList,
  ToolExecution,
  User,
} from "./types";
import { clearToken, getToken, setToken } from "./utils/token";

class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
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

export async function markChannelRead(
  channelId: number,
): Promise<{ last_read_message_id: number }> {
  return apiFetch<{ last_read_message_id: number }>(
    `/api/channels/${channelId}/read`,
    { method: "POST" },
  );
}

export async function getPersonaTemplates(): Promise<PersonaTemplate[]> {
  return apiFetch<PersonaTemplate[]>("/api/personas/templates");
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

export async function getAgentStatus(): Promise<AgentStatusResponse> {
  return apiFetch<AgentStatusResponse>("/api/agents/status");
}

export async function startAgents(): Promise<void> {
  return apiFetch<void>("/api/agents/start", { method: "POST" });
}

export async function stopAgents(): Promise<void> {
  return apiFetch<void>("/api/agents/stop", { method: "POST" });
}

export async function getAgentLLMCalls(
  personaId: number,
  opts?: { limit?: number; offset?: number },
): Promise<LLMCall[]> {
  const params = new URLSearchParams();
  if (opts?.limit) params.set("limit", String(opts.limit));
  if (opts?.offset) params.set("offset", String(opts.offset));
  const qs = params.toString();
  return apiFetch<LLMCall[]>(
    `/api/agents/${personaId}/llm-calls${qs ? `?${qs}` : ""}`,
  );
}

export async function getAgentToolExecutions(
  personaId: number,
  opts?: { limit?: number; offset?: number },
): Promise<ToolExecution[]> {
  const params = new URLSearchParams();
  if (opts?.limit) params.set("limit", String(opts.limit));
  if (opts?.offset) params.set("offset", String(opts.offset));
  const qs = params.toString();
  return apiFetch<ToolExecution[]>(
    `/api/agents/${personaId}/tool-executions${qs ? `?${qs}` : ""}`,
  );
}

export async function getAgentStats(
  personaId: number,
): Promise<AgentStatsResponse> {
  return apiFetch<AgentStatsResponse>(`/api/agents/${personaId}/stats`);
}

export async function addReaction(
  channelId: number,
  messageId: number,
  emoji: string,
): Promise<ReactionCount[]> {
  return apiFetch<ReactionCount[]>(
    `/api/channels/${channelId}/messages/${messageId}/reactions`,
    {
      method: "PUT",
      body: JSON.stringify({ emoji }),
    },
  );
}

export async function removeReaction(
  channelId: number,
  messageId: number,
  emoji: string,
): Promise<ReactionCount[]> {
  return apiFetch<ReactionCount[]>(
    `/api/channels/${channelId}/messages/${messageId}/reactions`,
    {
      method: "DELETE",
      body: JSON.stringify({ emoji }),
    },
  );
}

export async function listDMs(): Promise<DMChannel[]> {
  return apiFetch<DMChannel[]>("/api/dms");
}

export async function createDM(opts: {
  user_id?: number;
  persona_id?: number;
}): Promise<DMChannel> {
  return apiFetch<DMChannel>("/api/dms", {
    method: "POST",
    body: JSON.stringify(opts),
  });
}

export async function getUsers(): Promise<User[]> {
  return apiFetch<User[]>("/api/users");
}

export async function getChannelMembers(
  channelId: number,
): Promise<ChannelMember[]> {
  return apiFetch<ChannelMember[]>(`/api/channels/${channelId}/members`);
}

export async function addChannelMember(
  channelId: number,
  opts: { user_id?: number; persona_id?: number },
): Promise<void> {
  return apiFetch<void>(`/api/channels/${channelId}/members`, {
    method: "POST",
    body: JSON.stringify(opts),
  });
}

export async function removeChannelMember(
  channelId: number,
  opts: { user_id?: number; persona_id?: number },
): Promise<void> {
  return apiFetch<void>(`/api/channels/${channelId}/members`, {
    method: "DELETE",
    body: JSON.stringify(opts),
  });
}

// Projects

export async function getProjects(): Promise<Project[]> {
  return apiFetch<Project[]>("/api/projects");
}

export async function createProject(
  data: Omit<Project, "id" | "created_at">,
): Promise<Project> {
  return apiFetch<Project>("/api/projects", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateProject(
  id: number,
  data: Omit<Project, "id" | "created_at">,
): Promise<Project> {
  return apiFetch<Project>(`/api/projects/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export async function deleteProject(id: number): Promise<void> {
  return apiFetch<void>(`/api/projects/${id}`, { method: "DELETE" });
}

// Project Documents

export async function getProjectDocuments(
  projectId: number,
): Promise<ProjectDocumentList[]> {
  return apiFetch<ProjectDocumentList[]>(
    `/api/projects/${projectId}/documents`,
  );
}

export async function getCategoryDocuments(
  projectId: number,
  type: string,
): Promise<ProjectDocumentList> {
  return apiFetch<ProjectDocumentList>(
    `/api/projects/${projectId}/documents/${type}`,
  );
}

export async function getProjectDocument(
  projectId: number,
  type: string,
  filename: string,
): Promise<ProjectDocument> {
  return apiFetch<ProjectDocument>(
    `/api/projects/${projectId}/documents/${type}/${encodeURIComponent(filename)}`,
  );
}

export async function updateProjectDocument(
  projectId: number,
  type: string,
  filename: string,
  content: string,
): Promise<ProjectDocument> {
  return apiFetch<ProjectDocument>(
    `/api/projects/${projectId}/documents/${type}/${encodeURIComponent(filename)}`,
    {
      method: "PUT",
      body: JSON.stringify({ content }),
    },
  );
}

export async function appendToDocument(
  projectId: number,
  type: string,
  filename: string,
  content: string,
): Promise<void> {
  return apiFetch<void>(
    `/api/projects/${projectId}/documents/${type}/${encodeURIComponent(filename)}`,
    {
      method: "POST",
      body: JSON.stringify({ content }),
    },
  );
}

export async function deleteDocument(
  projectId: number,
  type: string,
  filename: string,
): Promise<void> {
  return apiFetch<void>(
    `/api/projects/${projectId}/documents/${type}/${encodeURIComponent(filename)}`,
    {
      method: "DELETE",
    },
  );
}

export async function getChannelProjects(
  channelId: number,
): Promise<Project[]> {
  return apiFetch<Project[]>(`/api/channels/${channelId}/projects`);
}

export async function addChannelProject(
  channelId: number,
  projectId: number,
): Promise<void> {
  return apiFetch<void>(`/api/channels/${channelId}/projects`, {
    method: "POST",
    body: JSON.stringify({ project_id: projectId }),
  });
}

export async function removeChannelProject(
  channelId: number,
  projectId: number,
): Promise<void> {
  return apiFetch<void>(`/api/channels/${channelId}/projects/${projectId}`, {
    method: "DELETE",
  });
}

export async function getMentionTargets(): Promise<MentionTarget[]> {
  return apiFetch<MentionTarget[]>("/api/mention-targets");
}

// Context budget

export async function getContextBudget(
  personaId: number,
  channelId: number,
): Promise<ContextBudget> {
  return apiFetch<ContextBudget>(
    `/api/agents/${personaId}/context-budget?channel_id=${channelId}`,
  );
}

export async function resetContext(
  personaId: number,
  channelId: number,
): Promise<void> {
  return apiFetch<void>(
    `/api/agents/${personaId}/channels/${channelId}/reset-context`,
    { method: "POST" },
  );
}

export { ApiError };
