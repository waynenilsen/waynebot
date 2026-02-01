# Better Markdown Support for Multiline Codeblocks

## Problem
The chat interface needs better markdown rendering support, specifically for multiline code blocks (triple backtick fenced blocks with language specifiers).

## What to do
1. Check `frontend/src/components/MarkdownRenderer.tsx` - audit current markdown rendering
2. Ensure fenced code blocks (```lang ... ```) render properly with:
   - Syntax highlighting (check if a library like highlight.js or prism is used)
   - Proper whitespace/indentation preservation
   - Language label display
   - Copy button (nice to have)
3. Test with various code block scenarios:
   - Multiple code blocks in one message
   - Nested backticks
   - Various languages (go, typescript, python, bash, etc.)
   - Very long code blocks

## Key files
- `frontend/src/components/MarkdownRenderer.tsx`
- `frontend/package.json` (may need to add syntax highlighting deps)
