import { useCallback, useEffect } from "react";
import * as api from "../api";
import { useApp } from "../store/AppContext";
import { useErrors } from "../store/ErrorContext";
import { getErrorMessage } from "../utils/errors";

export function useDMs() {
  const { state, setDMs, setCurrentChannel, clearDMUnread } = useApp();
  const { pushError } = useErrors();

  useEffect(() => {
    api
      .listDMs()
      .then(setDMs)
      .catch((err) => pushError(`Failed to load DMs: ${err.message}`));
  }, [setDMs, pushError]);

  const selectDM = useCallback(
    (id: number) => {
      setCurrentChannel(id);
      clearDMUnread(id);
      api.markChannelRead(id).catch(() => {});
    },
    [setCurrentChannel, clearDMUnread],
  );

  const createDM = useCallback(
    async (opts: { user_id?: number; persona_id?: number }) => {
      try {
        const dm = await api.createDM(opts);
        // Add to list if not already present
        const exists = state.dms.some((d) => d.id === dm.id);
        if (!exists) {
          setDMs([...state.dms, dm]);
        }
        setCurrentChannel(dm.id);
        return dm;
      } catch (err) {
        pushError(
          `Failed to create DM: ${getErrorMessage(err)}`,
        );
        throw err;
      }
    },
    [state.dms, setDMs, setCurrentChannel, pushError],
  );

  return {
    dms: state.dms,
    currentDM: state.dms.find((d) => d.id === state.currentChannelId) ?? null,
    selectDM,
    createDM,
  };
}
