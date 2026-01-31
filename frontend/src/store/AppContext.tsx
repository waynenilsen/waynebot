import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useReducer,
} from "react";
import type { ReactNode } from "react";
import type { Channel, Message, User } from "../types";

interface AppState {
  user: User | null;
  channels: Channel[];
  currentChannelId: number | null;
  messages: Record<number, Message[]>;
}

type AppAction =
  | { type: "SET_USER"; user: User | null }
  | { type: "SET_CHANNELS"; channels: Channel[] }
  | { type: "SET_CURRENT_CHANNEL"; channelId: number | null }
  | { type: "SET_MESSAGES"; channelId: number; messages: Message[] }
  | { type: "ADD_MESSAGE"; message: Message }
  | { type: "INCREMENT_UNREAD"; channelId: number }
  | { type: "CLEAR_UNREAD"; channelId: number };

function reducer(state: AppState, action: AppAction): AppState {
  switch (action.type) {
    case "SET_USER":
      return { ...state, user: action.user };
    case "SET_CHANNELS":
      return { ...state, channels: action.channels };
    case "SET_CURRENT_CHANNEL":
      return { ...state, currentChannelId: action.channelId };
    case "SET_MESSAGES":
      return {
        ...state,
        messages: { ...state.messages, [action.channelId]: action.messages },
      };
    case "ADD_MESSAGE": {
      const chId = action.message.channel_id;
      const existing = state.messages[chId] ?? [];
      if (existing.some((m) => m.id === action.message.id)) return state;
      return {
        ...state,
        messages: { ...state.messages, [chId]: [...existing, action.message] },
      };
    }
    case "INCREMENT_UNREAD":
      return {
        ...state,
        channels: state.channels.map((ch) =>
          ch.id === action.channelId
            ? { ...ch, unread_count: (ch.unread_count ?? 0) + 1 }
            : ch,
        ),
      };
    case "CLEAR_UNREAD":
      return {
        ...state,
        channels: state.channels.map((ch) =>
          ch.id === action.channelId ? { ...ch, unread_count: 0 } : ch,
        ),
      };
  }
}

interface AppContextValue {
  state: AppState;
  setUser: (user: User | null) => void;
  setChannels: (channels: Channel[]) => void;
  setCurrentChannel: (channelId: number | null) => void;
  setMessages: (channelId: number, messages: Message[]) => void;
  addMessage: (message: Message) => void;
  incrementUnread: (channelId: number) => void;
  clearUnread: (channelId: number) => void;
}

const AppContext = createContext<AppContextValue | null>(null);

const initialState: AppState = {
  user: null,
  channels: [],
  currentChannelId: null,
  messages: {},
};

export function AppProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(reducer, initialState);

  const setUser = useCallback(
    (user: User | null) => dispatch({ type: "SET_USER", user }),
    [],
  );
  const setChannels = useCallback(
    (channels: Channel[]) => dispatch({ type: "SET_CHANNELS", channels }),
    [],
  );
  const setCurrentChannel = useCallback(
    (channelId: number | null) =>
      dispatch({ type: "SET_CURRENT_CHANNEL", channelId }),
    [],
  );
  const setMessages = useCallback(
    (channelId: number, messages: Message[]) =>
      dispatch({ type: "SET_MESSAGES", channelId, messages }),
    [],
  );
  const addMessage = useCallback(
    (message: Message) => dispatch({ type: "ADD_MESSAGE", message }),
    [],
  );
  const incrementUnread = useCallback(
    (channelId: number) => dispatch({ type: "INCREMENT_UNREAD", channelId }),
    [],
  );
  const clearUnread = useCallback(
    (channelId: number) => dispatch({ type: "CLEAR_UNREAD", channelId }),
    [],
  );

  const value = useMemo(
    () => ({
      state,
      setUser,
      setChannels,
      setCurrentChannel,
      setMessages,
      addMessage,
      incrementUnread,
      clearUnread,
    }),
    [
      state,
      setUser,
      setChannels,
      setCurrentChannel,
      setMessages,
      addMessage,
      incrementUnread,
      clearUnread,
    ],
  );

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
}

export function useApp(): AppContextValue {
  const ctx = useContext(AppContext);
  if (!ctx) throw new Error("useApp must be used within AppProvider");
  return ctx;
}
