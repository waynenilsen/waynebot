import { useCallback, useEffect, useRef, useState } from "react";
import * as api from "../api";
import { useApp } from "../store/AppContext";

export function useMessages(channelId: number | null) {
  const { state, setMessages, addMessage } = useApp();
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
        setMessages(channelId, msgs);
        setHasMore(msgs.length >= 50);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [channelId, setMessages]);

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
        setMessages(channelId, [...older, ...messages]);
      }
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, [channelId, loading, hasMore, messages, setMessages]);

  const sendMessage = useCallback(
    async (content: string) => {
      if (!channelId) return;
      const msg = await api.postMessage(channelId, content);
      addMessage(msg);
    },
    [channelId, addMessage],
  );

  return { messages, loading, hasMore, loadMore, sendMessage };
}
