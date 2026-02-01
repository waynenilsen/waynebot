# Agent Memory and RAG

Add a memory system so agents always start by RAGging through previous messages and project context. Context window management is critical.

## Key Requirements

- **Memory store**: Agents need persistent memory — key facts, decisions, summaries from past conversations. must be a `memories` table with vector embeddings
- **RAG on startup**: Before responding, agents should retrieve relevant past messages and project context. This means searching previous messages by relevance (semantic or keyword), plus reading key project files if a project is associated.
- **Context window management**: This is critical. Track token usage carefully. The actor already has budget checking, but we need smarter context assembly:
  - Prioritize: system prompt → retrieved memories → recent channel messages → current message
  - do not Truncate or summarize older context when approaching limits
  - do not Consider a sliding window with summarization of older messages
  - summarization has shown itself to be problematic therefore, we must in stead start with fresh context often and deterministically compile relevant context the user must be aware of the remaining context for the agent and the user must take action on it the agent will go out of service when the context becomes full and the user must themselves take decisive action by ensuring the next context round has the appropriate contents
- **Project context**: If a channel has an associated project, include the AGENTS.md when present
- **Memory persistence**: After responding, agents should extract and store key facts/decisions for future retrieval.

agents will probably need a rag "tool" also

this should be couched in the language of software engineering as much as possible with specific carve outs for engineering requirements document, product requirements document, decision log, etcetera and loop it in with projects, and it should be file based as well
