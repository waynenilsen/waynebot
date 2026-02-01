# Direct Messages

Add DM support so users can have 1:1 conversations with agents (and other users).

## Key Requirements

- **DM channels**: A DM is a special channel type between exactly two participants (user-to-user or user-to-agent). Reuse the existing channel/message infrastructure but mark DMs distinctly (e.g. `is_dm` flag or `channel_type` column).
- **Backend**: New migration adding DM support to channels table. API endpoints to create/list DMs. When a DM is created with an agent persona, auto-subscribe that persona to the channel.
- **Frontend**: DM list in sidebar (separate from regular channels). DM compose/create flow. Integrate DM view into the agent activity screen so users can see agent thinking + inject messages right there.
- **Agent integration**: Agents should process DM channels the same way they process regular channels — the actor loop already watches subscribed channels, so subscribing the persona to the DM channel should be sufficient.

## Agent Activity Screen Integration

The existing agent activity viewer should be enhanced so that when viewing an agent's activity, the user can also see/send DMs to that agent inline. This turns the activity screen into a combined inspection + conversation view.
