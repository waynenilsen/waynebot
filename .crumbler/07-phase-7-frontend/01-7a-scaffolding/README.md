# 7a: Scaffolding

**Goal:** Set up Tailwind CSS, testing infrastructure, and Vite proxy config.

## Tasks

1. **Install Tailwind CSS v4** with `@tailwindcss/vite` plugin. Configure in `vite.config.ts`. Add `@import "tailwindcss"` to `src/index.css`. Remove the default Vite CSS boilerplate from `App.css` and `index.css`.

2. **Install testing deps:** `vitest`, `@testing-library/react`, `@testing-library/jest-dom`, `@testing-library/user-event`, `jsdom`. Add `vitest.config.ts` (or configure in `vite.config.ts`) with `environment: 'jsdom'`. Add `src/test/setup.ts` that imports `@testing-library/jest-dom`. Add `"test": "vitest run"` script to package.json.

3. **Configure Vite proxy** in `vite.config.ts`:
   - `/api` → `http://localhost:59731`
   - `/ws` → `ws://localhost:59731` (WebSocket upgrade)
   - Dev server port: `53461` (already set)

4. **Update CORS origin** — The Go backend defaults to `http://localhost:5173`. Set env var `WAYNEBOT_CORS_ORIGINS=http://localhost:53461` or update the proxy to handle it. Since we're proxying through Vite, CORS shouldn't matter (same origin). No backend change needed.

5. **Write a smoke test** `src/App.test.tsx` that renders `<App />` and asserts it mounts without crashing.

## Verification

- `cd frontend && npx vitest run` passes the smoke test.
- `npm run build` succeeds with no TS errors.
- Tailwind utility classes work (verify by adding a class to App.tsx temporarily).
