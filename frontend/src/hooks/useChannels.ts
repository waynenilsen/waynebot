import { useCallback, useEffect } from "react";
import * as api from "../api";
import { useApp } from "../store/AppContext";

export function useChannels() {
  const { state, setChannels, setCurrentChannel } = useApp();

  useEffect(() => {
    api.getChannels().then(setChannels).catch(() => {});
  }, [setChannels]);

  const selectChannel = useCallback(
    (id: number) => {
      setCurrentChannel(id);
    },
    [setCurrentChannel],
  );

  const createChannel = useCallback(
    async (name: string, description: string) => {
      const ch = await api.createChannel(name, description);
      setChannels([...state.channels, ch]);
      return ch;
    },
    [state.channels, setChannels],
  );

  return {
    channels: state.channels,
    currentChannel: state.channels.find(
      (c) => c.id === state.currentChannelId,
    ) ?? null,
    currentChannelId: state.currentChannelId,
    selectChannel,
    createChannel,
  };
}
