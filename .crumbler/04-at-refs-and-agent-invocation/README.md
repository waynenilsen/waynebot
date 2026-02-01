# @Refs and Agent Invocation

Add @mention support so users can ping agents (and other users) in messages.

## Key Requirements

- **Parsing**: Parse `@name` references in message content. Match against known usernames and persona names.
- **Backend**: When a message contains `@persona_name`, ensure that persona is notified even if it wasn't already watching the channel. This could mean: auto-subscribing the persona to the channel, or having a separate notification mechanism. The decision maker should treat an @mention as a strong signal to respond.
- **Frontend**: Autocomplete dropdown when typing `@` in the compose box. Show @mentions with distinct styling (highlight/link). List of available @targets from users + personas.
- **Agent decision making**: Update the `DecisionMaker` to always respond when directly @mentioned, regardless of other heuristics.
