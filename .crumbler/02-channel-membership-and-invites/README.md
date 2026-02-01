# Channel Membership & Invites

Right now every human sees every channel and there's no concept of channel membership. Agents have a `persona_channels` subscription table but there's no UX to manage it — you have to manually insert rows in the DB.

This needs a unified channel membership system that covers both humans and bots.

## The Problem

- **Humans**: No membership model at all. All channels are visible to all users. There's no way to have private channels or control who sees what.
- **Agents**: The `persona_channels` table and `SubscribeChannel`/`UnsubscribeChannel` model functions exist, but no API endpoints or UI expose them. Agents won't respond to anything until someone manually inserts a row.
- **No invite flow**: There's no way to invite a user or bot into a channel. This is the natural UX for both — you create a channel, then invite participants (humans or agents) into it.

## Key Requirements

- **Channel membership table**: A unified `channel_members` table (or extend `persona_channels`) that tracks both human and agent membership per channel. Channel creators are auto-added.
- **Invite UX**: A way to invite users and agents into a channel. Could be a `/invite @name` command, a UI button, or both. Inviting an agent should auto-subscribe its persona to the channel.
- **Visibility**: Users should only see channels they're members of. Agents should only watch channels they're subscribed to (this part already works).
- **API endpoints**: Wire up subscribe/unsubscribe for both humans and agents. The agent model functions already exist, just need endpoints and equivalent human membership logic.
