import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import ChannelList from "./ChannelList";
import type { Channel } from "../types";

const general: Channel = {
  id: 1,
  name: "general",
  description: "General chat",
  created_at: "2024-01-01T00:00:00Z",
  unread_count: 0,
};

const random: Channel = {
  id: 2,
  name: "random",
  description: "Random stuff",
  created_at: "2024-01-01T00:00:00Z",
  unread_count: 0,
};

function scenario(overrides?: {
  channels?: Channel[];
  currentChannelId?: number | null;
  onSelect?: (id: number) => void;
  onCreate?: (name: string, description: string) => Promise<void>;
}) {
  const onSelect = overrides?.onSelect ?? vi.fn();
  const onCreate = overrides?.onCreate ?? vi.fn(() => Promise.resolve());
  const user = userEvent.setup();

  render(
    <ChannelList
      channels={overrides?.channels ?? [general, random]}
      currentChannelId={overrides?.currentChannelId ?? null}
      onSelect={onSelect}
      onCreate={onCreate}
    />,
  );

  return {
    user,
    onSelect: onSelect as ReturnType<typeof vi.fn>,
    onCreate: onCreate as ReturnType<typeof vi.fn>,
  };
}

describe("ChannelList", () => {
  it("renders channels with # prefix", () => {
    scenario();
    expect(screen.getByText("general")).toBeInTheDocument();
    expect(screen.getByText("random")).toBeInTheDocument();
    expect(screen.getAllByText("#")).toHaveLength(2);
  });

  it("highlights the active channel", () => {
    scenario({ currentChannelId: 1 });
    const generalBtn = screen.getByText("general").closest("button")!;
    expect(generalBtn.className).toContain("text-[#e2b714]");
  });

  it("calls onSelect when a channel is clicked", async () => {
    const s = scenario();
    await s.user.click(screen.getByText("general"));
    expect(s.onSelect).toHaveBeenCalledWith(1);
  });

  it("shows empty state when no channels", () => {
    scenario({ channels: [] });
    expect(screen.getByText("no channels yet")).toBeInTheDocument();
  });

  it("toggles create form on + click", async () => {
    const s = scenario();
    expect(
      screen.queryByPlaceholderText("channel-name"),
    ).not.toBeInTheDocument();

    await s.user.click(screen.getByTitle("New channel"));
    expect(screen.getByPlaceholderText("channel-name")).toBeInTheDocument();
  });

  it("creates a channel via the form", async () => {
    const s = scenario();

    await s.user.click(screen.getByTitle("New channel"));
    await s.user.type(
      screen.getByPlaceholderText("channel-name"),
      "new-channel",
    );
    await s.user.type(
      screen.getByPlaceholderText("description (optional)"),
      "A description",
    );
    await s.user.click(screen.getByText("Create"));

    expect(s.onCreate).toHaveBeenCalledWith("new-channel", "A description");
  });

  it("closes create form after successful creation", async () => {
    const s = scenario();

    await s.user.click(screen.getByTitle("New channel"));
    await s.user.type(screen.getByPlaceholderText("channel-name"), "test");
    await s.user.click(screen.getByText("Create"));

    expect(
      screen.queryByPlaceholderText("channel-name"),
    ).not.toBeInTheDocument();
  });

  it("disables submit when name is empty", async () => {
    const s = scenario();
    await s.user.click(screen.getByTitle("New channel"));

    const createBtn = screen.getByText("Create");
    expect(createBtn).toBeDisabled();
  });
});
