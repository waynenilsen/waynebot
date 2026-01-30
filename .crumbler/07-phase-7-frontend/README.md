# Phase 7: Frontend

**Goal:** Working Slack-clone UI with high design quality. All UI components must be built using the `/frontend-design` skill.

## Sub-phases

This phase should be decomposed into:

- **7a: Scaffolding** — Add Tailwind, Vitest + React Testing Library + jsdom. Configure vite.config.ts to proxy `/api` and `/ws` to Go backend (port 59731). Dev server runs on port 53461.
- **7b: Types + API + WebSocket + Store** — types.ts, api.ts, ws.ts, store/AppContext.tsx, utils/. Tests for api.ts, ws.ts, AppContext.tsx.
- **7c: Auth** — useAuth.ts, LoginPage.tsx (use /frontend-design), auth gate in App.tsx. Tests.
- **7d: Layout + Channels** — Layout.tsx, Sidebar.tsx, ChannelList.tsx, useChannels.ts, useWebSocket.ts (use /frontend-design). Tests.
- **7e: Messages** — MessageThread.tsx, MessageItem.tsx, MessageCompose.tsx, MarkdownRenderer.tsx, useMessages.ts (use /frontend-design). Tests.
- **7f: Persona + Agent + Invite pages** — PersonaPage.tsx, PersonaForm.tsx, AgentDashboard.tsx, InvitePage.tsx + hooks (use /frontend-design).
- **7g: Polish** — ErrorBanner.tsx, empty states, loading spinners, WebSocket reconnection indicator, keyboard shortcuts.

## Verification

`npx vitest run` passes. Full flow in browser works.
