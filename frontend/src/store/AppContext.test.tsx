import { render, screen, act } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { AppProvider, useApp } from "./AppContext";
import type { Channel, Message, User } from "../types";

function scenario() {
  let ctx: ReturnType<typeof useApp>;

  function TestConsumer() {
    ctx = useApp();
    return (
      <div>
        <span data-testid="user">{ctx.state.user?.username ?? "none"}</span>
        <span data-testid="channels">{ctx.state.channels.length}</span>
        <span data-testid="currentChannel">
          {ctx.state.currentChannelId ?? "none"}
        </span>
      </div>
    );
  }

  render(
    <AppProvider>
      <TestConsumer />
    </AppProvider>,
  );

  return {
    getCtx: () => ctx,
    getText: (testId: string) => screen.getByTestId(testId).textContent,
  };
}

const alice: User = {
  id: 1,
  username: "alice",
  created_at: "2024-01-01T00:00:00Z",
};

const general: Channel = {
  id: 1,
  name: "general",
  description: "General chat",
  created_at: "2024-01-01T00:00:00Z",
  unread_count: 0,
};

const msg: Message = {
  id: 10,
  channel_id: 1,
  author_id: 1,
  author_type: "human",
  author_name: "alice",
  content: "hello world",
  created_at: "2024-01-01T00:00:00Z",
  reactions: null,
};

describe("AppContext", () => {
  it("starts with null user and empty state", () => {
    const { getText } = scenario();
    expect(getText("user")).toBe("none");
    expect(getText("channels")).toBe("0");
    expect(getText("currentChannel")).toBe("none");
  });

  it("setUser updates user state", () => {
    const { getCtx, getText } = scenario();
    act(() => getCtx().setUser(alice));
    expect(getText("user")).toBe("alice");
  });

  it("setChannels updates channels", () => {
    const { getCtx, getText } = scenario();
    act(() => getCtx().setChannels([general]));
    expect(getText("channels")).toBe("1");
  });

  it("setCurrentChannel updates current channel id", () => {
    const { getCtx, getText } = scenario();
    act(() => getCtx().setCurrentChannel(1));
    expect(getText("currentChannel")).toBe("1");
  });

  it("setMessages stores messages for a channel", () => {
    const { getCtx } = scenario();
    act(() => getCtx().setMessages(1, [msg]));
    expect(getCtx().state.messages[1]).toEqual([msg]);
  });

  it("addMessage appends to existing channel messages", () => {
    const { getCtx } = scenario();
    act(() => getCtx().setMessages(1, [msg]));

    const msg2: Message = { ...msg, id: 11, content: "second" };
    act(() => getCtx().addMessage(msg2));

    expect(getCtx().state.messages[1]).toHaveLength(2);
    expect(getCtx().state.messages[1][1].content).toBe("second");
  });

  it("addMessage creates channel entry if none exists", () => {
    const { getCtx } = scenario();
    const msg3: Message = { ...msg, channel_id: 99 };
    act(() => getCtx().addMessage(msg3));
    expect(getCtx().state.messages[99]).toHaveLength(1);
  });

  it("throws when useApp is used outside provider", () => {
    function BadConsumer() {
      useApp();
      return null;
    }
    expect(() => render(<BadConsumer />)).toThrow(
      "useApp must be used within AppProvider",
    );
  });
});
