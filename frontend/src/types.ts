export interface User {
  id: number;
  username: string;
  created_at: string;
}

export interface Channel {
  id: number;
  name: string;
  description: string;
  created_at: string;
}

export interface Message {
  id: number;
  channel_id: number;
  author_id: number;
  author_type: "human" | "agent";
  author_name: string;
  content: string;
  created_at: string;
}

export interface Persona {
  id: number;
  name: string;
  system_prompt: string;
  model: string;
  tools_enabled: string[];
  temperature: number;
  max_tokens: number;
  cooldown_secs: number;
  max_tokens_per_hour: number;
  created_at: string;
}

export interface Invite {
  id: number;
  code: string;
  created_by: number;
  used_by: number | null;
  created_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface AgentStatus {
  persona_id: number;
  persona_name: string;
  status: string;
  channels: string[];
}

export interface WsEvent {
  type: string;
  data: unknown;
}
