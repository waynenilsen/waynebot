# Agent Memory and RAG

Add a memory system so agents always start by RAGging through previous messages and project context. Context window management is critical.

## Key Requirements

- **Memory store**: Agents need persistent memory — key facts, decisions, summaries from past conversations. Could be a `memories` table with vector embeddings, or a simpler keyword/tag-based retrieval system. Start simple (keyword/recent message retrieval) and iterate.
- **RAG on startup**: Before responding, agents should retrieve relevant past messages and project context. This means searching previous messages by relevance (semantic or keyword), plus reading key project files if a project is associated.
- **Context window management**: This is critical. Track token usage carefully. The actor already has budget checking, but we need smarter context assembly:
  - Prioritize: system prompt → retrieved memories → recent channel messages → current message
  - Truncate or summarize older context when approaching limits
  - Consider a sliding window with summarization of older messages
- **Project context**: If a channel has an associated project, include relevant project files in the context (e.g. README, key source files). Use the project path from the projects feature.
- **Memory persistence**: After responding, agents should extract and store key facts/decisions for future retrieval.
