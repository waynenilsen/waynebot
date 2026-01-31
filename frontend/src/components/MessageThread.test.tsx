import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import MessageThread from "./MessageThread";
import type { Message } from "../types";

vi.mock("./MessageItem", () => ({
  default: ({ message }: { message: Message }) => (
    <div data-testid={`msg-${message.id}`}>{message.content}</div>
  ),
}));

function msg(id: number, content = `message ${id}`): Message {
  return {
    id,
    channel_id: 1,
    author_id: 1,
    author_type: "human",
    author_name: "alice",
    content,
    created_at: "2024-01-01T00:00:00Z",
  };
}

function scenario(overrides?: {
  messages?: Message[];
  loading?: boolean;
  hasMore?: boolean;
  onLoadMore?: () => void;
  channelName?: string;
}) {
  const onLoadMore = overrides?.onLoadMore ?? vi.fn();
  const user = userEvent.setup();

  render(
    <MessageThread
      messages={overrides?.messages ?? [msg(1), msg(2)]}
      loading={overrides?.loading ?? false}
      hasMore={overrides?.hasMore ?? false}
      onLoadMore={onLoadMore}
      channelName={overrides?.channelName ?? "general"}
    />,
  );

  return { user, onLoadMore: onLoadMore as ReturnType<typeof vi.fn> };
}

describe("MessageThread", () => {
  it("renders channel name in header", () => {
    scenario({ channelName: "help" });
    expect(screen.getByText("help")).toBeInTheDocument();
    expect(screen.getByText("#")).toBeInTheDocument();
  });

  it("renders all messages", () => {
    scenario({ messages: [msg(1), msg(2), msg(3)] });
    expect(screen.getByTestId("msg-1")).toBeInTheDocument();
    expect(screen.getByTestId("msg-2")).toBeInTheDocument();
    expect(screen.getByTestId("msg-3")).toBeInTheDocument();
  });

  it("shows empty state when no messages and not loading", () => {
    scenario({ messages: [], loading: false });
    expect(screen.getByText(/no messages yet/)).toBeInTheDocument();
  });

  it("shows loading state when loading with no messages", () => {
    scenario({ messages: [], loading: true });
    expect(screen.getByText("loading...")).toBeInTheDocument();
  });

  it("shows load more button when hasMore is true", () => {
    scenario({ hasMore: true });
    expect(
      screen.getByText("load older messages"),
    ).toBeInTheDocument();
  });

  it("calls onLoadMore when load more button is clicked", async () => {
    const s = scenario({ hasMore: true });
    await s.user.click(screen.getByText("load older messages"));
    expect(s.onLoadMore).toHaveBeenCalledOnce();
  });

  it("disables load more button while loading", () => {
    scenario({ hasMore: true, loading: true });
    const btn = screen.getByText("loading...");
    expect(btn).toBeDisabled();
  });

  it("does not show load more when hasMore is false", () => {
    scenario({ hasMore: false });
    expect(
      screen.queryByText("load older messages"),
    ).not.toBeInTheDocument();
  });
});
