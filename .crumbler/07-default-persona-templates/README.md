# Default Persona Templates

## Problem
The agent needs opinionated default prompts to help it work effectively. Can't lean on the user to write good system prompts. All templates are around building software. Backend is Go, frontend is Vite+React+TS SPA.

## Templates to create

### 1. Code Architect
- Defines interfaces and documentation
- API design
- DB design
- Seams and code organization
- Opinionated about clean architecture, separation of concerns
- Reviews and critiques designs before implementation

### 2. Senior Backend Engineer
- Go expert
- API implementation, DB queries, business logic
- Performance-minded, idiomatic Go
- Testing-focused

### 3. Senior Frontend Engineer
- React/TypeScript/Vite expert
- Component design, state management, hooks
- Accessibility, responsive design
- Testing with Vitest/RTL

### 4. Senior QA Engineer
- Test strategy and planning
- Edge cases, regression testing
- Integration and E2E testing
- Bug reproduction and documentation

### 5. Product Manager
- Requirements gathering and refinement
- User story writing
- Feature prioritization
- Acceptance criteria definition

## What to do
1. Create these as selectable templates in the persona creation flow
2. Store templates as constants or a config file in the backend
3. Frontend: Add a template selector dropdown/modal when creating a new persona that pre-fills the system prompt
4. Each template should be a well-crafted, opinionated system prompt (not generic)
5. Templates should reference the specific tech stack (Go, React, TS, Vite, SQLite)

## Key files
- New file: `internal/agent/templates.go` or similar
- `internal/api/persona_handlers.go` - add endpoint to list templates
- `frontend/src/pages/PersonaPage.tsx` - add template selector UI
- `frontend/src/api.ts` - add API call for templates
