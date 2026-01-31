# 002: Persona page is not scrollable

## Symptom

The Edit Persona form content is cut off at the bottom of the viewport. Fields like Cooldown, Max Tokens/Hour, Tools Enabled, and the submit/cancel buttons are unreachable.

## Root cause

In `frontend/src/pages/PersonaPage.tsx:51`, the outer container uses `overflow-hidden`:

```tsx
<div className="flex-1 flex flex-col overflow-hidden">
```

This prevents the inner content from scrolling. The form view (line 35) correctly uses `overflow-y-auto`, but the list/edit wrapper does not.

## Fix

Change `overflow-hidden` to `overflow-y-auto` on line 51 to match the pattern used by the form view on line 35.

## Files involved

| File | Line | Role |
|------|------|------|
| `frontend/src/pages/PersonaPage.tsx` | 51 | `overflow-hidden` blocks scroll |
| `frontend/src/pages/PersonaPage.tsx` | 35 | working reference (`overflow-y-auto`) |
