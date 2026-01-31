# 7c: Auth

**Goal:** Login/register page and auth gate. Use `/frontend-design` skill for all UI components.

## Files to Create

### `src/hooks/useAuth.ts`
Custom hook providing:
- `user: User | null`
- `loading: boolean` (true while checking `/api/auth/me` on mount)
- `login(username, password)` — calls api.login, stores token, sets user in context
- `register(username, password, inviteCode?)` — calls api.register, stores token, sets user
- `logout()` — calls api.logout, clears token, clears user

On mount, if token exists in localStorage, call `getMe()` to validate session. If it fails (401), clear token.

### `src/pages/LoginPage.tsx`
**Use `/frontend-design` skill to build this.**

Slack-inspired login/register page with:
- Toggle between Login and Register modes
- Username field (alphanumeric + underscore, 1-50 chars)
- Password field (8-128 chars)
- Invite code field (only in Register mode, optional for first user)
- Submit button
- Error display for failed attempts
- Clean, modern design matching a Slack-clone aesthetic

### `src/App.tsx`
Auth gate:
- If `loading`, show a loading spinner/skeleton
- If no `user`, show `<LoginPage />`
- If `user`, show the main app layout (placeholder for now, e.g. "Welcome, {username}")

## Tests

- `src/hooks/useAuth.test.ts` — mock api calls, test login/register/logout flows, test session restoration on mount
- `src/pages/LoginPage.test.tsx` — render login page, fill form, submit, assert api called. Test toggle between login/register. Test error display.

## Verification

`npx vitest run` passes. Login flow works in browser against running backend.
