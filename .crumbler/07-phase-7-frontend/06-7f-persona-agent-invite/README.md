# 7f: Persona + Agent + Invite Pages

**Goal:** Admin pages for managing personas, viewing agent status, and generating invites. Use `/frontend-design` skill for all UI components.

## Files to Create

### `src/hooks/usePersonas.ts`
- CRUD operations: `personas`, `createPersona(data)`, `updatePersona(id, data)`, `deletePersona(id)`
- Fetch on mount

### `src/hooks/useAgents.ts`
- `agents: AgentStatus[]`, `startAgents()`, `stopAgents()`, `refreshStatus()`
- Poll status or refresh on demand

### `src/hooks/useInvites.ts`
- `invites: Invite[]`, `createInvite()`
- Fetch on mount

### `src/pages/PersonaPage.tsx`
**Use `/frontend-design` skill.**

- List all personas in a table/card layout
- Click to edit, button to create new
- Delete with confirmation

### `src/components/PersonaForm.tsx`
**Use `/frontend-design` skill.**

- Form for create/edit persona with fields:
  - Name (1-100 chars)
  - System prompt (textarea, 1-50000 chars)
  - Model (text input or dropdown of known models)
  - Temperature (0-2 slider/input)
  - Max tokens (number input)
  - Cooldown seconds (number input)
  - Max tokens per hour (number input)
  - Tools enabled (multi-select or tag input)
- Validation matching backend rules

### `src/pages/AgentDashboard.tsx`
**Use `/frontend-design` skill.**

- Show all agent statuses in a dashboard view
- Each agent: persona name, status (running/stopped), channels
- Global start/stop buttons
- Auto-refresh or manual refresh button

### `src/pages/InvitePage.tsx`
**Use `/frontend-design` skill.**

- List existing invites with status (used/unused, who used it)
- "Generate Invite" button
- Copy invite code to clipboard

### Navigation
- Add routing between channel view, PersonaPage, AgentDashboard, InvitePage
- Links in sidebar (from 7d)

## Tests

- `src/hooks/usePersonas.test.ts` — mock api, test CRUD
- `src/hooks/useInvites.test.ts` — mock api, test create and list
- `src/pages/PersonaPage.test.tsx` — render, test list display, test delete confirmation
- `src/pages/InvitePage.test.tsx` — render, test generate and copy

## Verification

`npx vitest run` passes. All admin pages render and function in browser.
