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
  unread_count: number;
}

export interface ReactionCount {
  emoji: string;
  count: number;
  reacted: boolean;
}

export interface Message {
  id: number;
  channel_id: number;
  author_id: number;
  author_type: "human" | "agent";
  author_name: string;
  content: string;
  created_at: string;
  reactions: ReactionCount[] | null;
}

export interface ReactionEvent {
  message_id: number;
  channel_id: number;
  emoji: string;
  author_id: number;
  author_type: string;
  counts: ReactionCount[];
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

export interface AgentStatusResponse {
  supervisor_running: boolean;
  agents: AgentStatus[];
}

export interface WsEvent {
  type: string;
  data: unknown;
}

export interface LLMCall {
  id: number;
  persona_id: number;
  channel_id: number;
  model: string;
  messages_json: string;
  response_json: string;
  prompt_tokens: number;
  completion_tokens: number;
  created_at: string;
}

export interface ToolExecution {
  id: number;
  persona_id: number;
  tool_name: string;
  args_json: string;
  output_text: string;
  error_text: string;
  duration_ms: number;
  created_at: string;
}

export interface DMChannel {
  id: number;
  name: string;
  is_dm: true;
  created_at: string;
  other_participant: {
    user_id: number | null;
    user_name: string | null;
    persona_id: number | null;
    persona_name: string | null;
  };
  unread_count: number;
}

export interface AgentStatsResponse {
  total_calls_last_hour: number;
  total_tokens_last_hour: number;
  error_count_last_hour: number;
  avg_response_ms: number;
}

export interface MentionTarget {
  type: "user" | "persona";
  id: number;
  name: string;
}

export interface ChannelMember {
  type: "user" | "persona";
  id: number;
  name: string;
  role: string;
}

export interface Project {
  id: number;
  name: string;
  path: string;
  description: string;
  created_at: string;
}

export interface ProjectDocument {
  type: string;
  content: string;
  exists: boolean;
}

export interface ContextBudget {
  persona_id: number;
  channel_id: number;
  total_tokens: number;
  system_tokens: number;
  project_tokens: number;
  memory_tokens: number;
  history_tokens: number;
  history_messages: number;
  exhausted: boolean;
}
