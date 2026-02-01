import { useCallback, useEffect, useRef, useState } from "react";
import * as api from "../api";
import { useApp } from "../store/AppContext";
import { useErrors } from "../store/ErrorContext";

export function useMessages(channelId: number | null) {
  const { state, setMessages, addMessage, updateReactions } = useApp();
  const { pushError } = useErrors();
  const [loading, setLoading] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const loadedRef = useRef<Set<number>>(new Set());

  const messages = channelId ? (state.messages[channelId] ?? []) : [];

  useEffect(() => {
    if (!channelId || loadedRef.current.has(channelId)) return;

    loadedRef.current.add(channelId);
    setLoading(true);
    setHasMore(true);

    api
      .getMessages(channelId, { limit: 50 })
      .then((msgs) => {
        // Backend returns newest-first, but we need oldest-first for display
        const reversed = [...msgs].reverse();
        setMessages(channelId, reversed);
        setHasMore(msgs.length >= 50);
      })
      .catch((err) => pushError(`Failed to load messages: ${err.message}`))
      .finally(() => setLoading(false));
  }, [channelId, setMessages, pushError]);

  const loadMore = useCallback(async () => {
    if (!channelId || loading || !hasMore || messages.length === 0) return;

    const oldest = messages[0];
    setLoading(true);
    try {
      const older = await api.getMessages(channelId, {
        limit: 50,
        before: oldest.id,
      });
      if (older.length < 50) setHasMore(false);
      if (older.length > 0) {
        // Backend returns newest-first, reverse to get oldest-first
        const reversed = [...older].reverse();
        setMessages(channelId, [...reversed, ...messages]);
      }
    } catch (err) {
      pushError(
        `Failed to load older messages: ${err instanceof Error ? err.message : "unknown error"}`,
      );
    } finally {
      setLoading(false);
    }
  }, [channelId, loading, hasMore, messages, setMessages, pushError]);

  const sendMessage = useCallback(
    async (content: string) => {
      if (!channelId) return;
      try {
        const msg = await api.postMessage(channelId, content);
        addMessage(msg);
      } catch (err) {
        pushError(
          `Failed to send message: ${err instanceof Error ? err.message : "unknown error"}`,
        );
        throw err;
      }
    },
    [channelId, addMessage, pushError],
  );

  const toggleReaction = useCallback(
    async (messageId: number, emoji: string, reacted: boolean) => {
      if (!channelId) return;
      try {
        const counts = reacted
          ? await api.removeReaction(channelId, messageId, emoji)
          : await api.addReaction(channelId, messageId, emoji);
        updateReactions(channelId, messageId, counts);
      } catch (err) {
        pushError(
          `Failed to toggle reaction: ${err instanceof Error ? err.message : "unknown error"}`,
        );
      }
    },
    [channelId, updateReactions, pushError],
  );

  return { messages, loading, hasMore, loadMore, sendMessage, toggleReaction };
}
