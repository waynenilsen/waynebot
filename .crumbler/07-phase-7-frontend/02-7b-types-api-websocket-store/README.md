# 7b: Types + API + WebSocket + Store

**Goal:** Build the data layer — TypeScript types matching the Go backend, API client, WebSocket client, and React context store.

## Backend API Reference

- Base URL: proxied via Vite, so just use relative paths like `/api/...`
- Auth: `Authorization: Bearer <token>` header or `session` cookie
- All errors: `{ "error": "message" }`

## Files to Create

### `src/types.ts`
TypeScript interfaces matching Go models:
```ts
interface User { id: number; username: string; created_at: string }
interface Channel { id: number; name: string; description: string; created_at: string }
interface Message { id: number; channel_id: number; author_id: number; author_type: 'human' | 'persona'; author_name: string; content: string; created_at: string }
interface Persona { id: number; name: string; system_prompt: string; model: string; tools_enabled: string[]; temperature: number; max_tokens: number; cooldown_secs: number; max_tokens_per_hour: number; created_at: string }
interface Invite { id: number; code: string; created_by: number; used_by: number | null; created_at: string }
interface AuthResponse { token: string; user: User }
interface AgentStatus { persona_id: number; persona_name: string; status: string; channels: string[] }
interface WsEvent { type: string; data: unknown }
```

### `src/api.ts`
Functions wrapping fetch calls. Store token in memory (or localStorage). Include:
- `register(username, password, inviteCode?)` → AuthResponse
- `login(username, password)` → AuthResponse
- `logout()`
- `getMe()` → User
- `getChannels()` → Channel[]
- `createChannel(name, description)` → Channel
- `getMessages(channelId, opts?: { limit?, before? })` → Message[]
- `postMessage(channelId, content)` → Message
- `getPersonas()` → Persona[]
- `createPersona(data)` → Persona
- `updatePersona(id, data)` → Persona
- `deletePersona(id)`
- `getInvites()` → Invite[]
- `createInvite()` → Invite
- `getAgentStatus()` → AgentStatus[]
- `startAgents()`
- `stopAgents()`

Use a helper like `apiFetch(path, opts)` that adds auth header and parses JSON. Throw on non-ok responses with the error message from the body.

### `src/ws.ts`
WebSocket client:
- `connectWs(onEvent: (event: WsEvent) => void): { close: () => void }`
- First fetch `/api/ws/ticket` to get a ticket, then connect to `ws://HOST/ws?ticket=...`
- Use relative URL construction for WebSocket (derive from `window.location`)
- Auto-reconnect with exponential backoff
- Expose connection state for UI indicators

### `src/store/AppContext.tsx`
React context + provider with state:
- `user: User | null`
- `channels: Channel[]`
- `currentChannelId: number | null`
- `messages: Record<number, Message[]>` (keyed by channel ID)
- Actions: `setUser`, `setChannels`, `setCurrentChannel`, `addMessage`, `setMessages`
- On `new_message` WS event, append to the right channel's message list

### `src/utils/token.ts`
Simple token storage: `getToken()`, `setToken(t)`, `clearToken()` using localStorage.

## Tests

- `src/api.test.ts` — mock fetch, test that `login` sends correct body and returns parsed response, test error handling
- `src/ws.test.ts` — mock WebSocket, test ticket fetch + connection + event dispatch
- `src/store/AppContext.test.tsx` — render provider, dispatch actions, assert state changes

## Verification

`npx vitest run` passes all tests.
