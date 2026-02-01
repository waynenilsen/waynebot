import { useCallback, useEffect } from "react";
import * as api from "../api";
import { useApp } from "../store/AppContext";
import { useErrors } from "../store/ErrorContext";
import { getErrorMessage } from "../utils/errors";

export function useChannels() {
  const { state, setChannels, setCurrentChannel, clearUnread } = useApp();
  const { pushError } = useErrors();

  useEffect(() => {
    api
      .getChannels()
      .then(setChannels)
      .catch((err) => pushError(`Failed to load channels: ${err.message}`));
  }, [setChannels, pushError]);

  const selectChannel = useCallback(
    (id: number) => {
      setCurrentChannel(id);
      clearUnread(id);
      // Fire-and-forget: tell the server we've read this channel.
      api.markChannelRead(id).catch(() => {});
    },
    [setCurrentChannel, clearUnread],
  );

  const createChannel = useCallback(
    async (name: string, description: string) => {
      try {
        const ch = await api.createChannel(name, description);
        setChannels([...state.channels, ch]);
        return ch;
      } catch (err) {
        pushError(
          `Failed to create channel: ${getErrorMessage(err)}`,
        );
        throw err;
      }
    },
    [state.channels, setChannels, pushError],
  );

  return {
    channels: state.channels,
    currentChannel:
      state.channels.find((c) => c.id === state.currentChannelId) ?? null,
    currentChannelId: state.currentChannelId,
    selectChannel,
    createChannel,
  };
}
