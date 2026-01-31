import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import MessageCompose from "./MessageCompose";

function scenario(overrides?: {
  onSend?: (content: string) => Promise<void>;
}) {
  const onSend =
    overrides?.onSend ?? vi.fn(() => Promise.resolve());
  const user = userEvent.setup();

  render(<MessageCompose onSend={onSend} />);

  return { user, onSend: onSend as ReturnType<typeof vi.fn> };
}

describe("MessageCompose", () => {
  it("renders textarea and send button", () => {
    scenario();
    expect(
      screen.getByPlaceholderText("type a message..."),
    ).toBeInTheDocument();
    expect(screen.getByText("send")).toBeInTheDocument();
  });

  it("send button is disabled when textarea is empty", () => {
    scenario();
    expect(screen.getByText("send")).toBeDisabled();
  });

  it("send button enables when text is entered", async () => {
    const s = scenario();
    await s.user.type(
      screen.getByPlaceholderText("type a message..."),
      "hello",
    );
    expect(screen.getByText("send")).toBeEnabled();
  });

  it("calls onSend with trimmed content on button click", async () => {
    const s = scenario();
    await s.user.type(
      screen.getByPlaceholderText("type a message..."),
      "  hello world  ",
    );
    await s.user.click(screen.getByText("send"));
    expect(s.onSend).toHaveBeenCalledWith("hello world");
  });

  it("clears textarea after successful send", async () => {
    const s = scenario();
    const textarea = screen.getByPlaceholderText("type a message...");
    await s.user.type(textarea, "hello");
    await s.user.click(screen.getByText("send"));
    expect(textarea).toHaveValue("");
  });

  it("sends on Enter key", async () => {
    const s = scenario();
    const textarea = screen.getByPlaceholderText("type a message...");
    await s.user.type(textarea, "hello{Enter}");
    expect(s.onSend).toHaveBeenCalledWith("hello");
  });

  it("does not send on Shift+Enter (allows newline)", async () => {
    const s = scenario();
    const textarea = screen.getByPlaceholderText("type a message...");
    await s.user.type(textarea, "hello{Shift>}{Enter}{/Shift}");
    expect(s.onSend).not.toHaveBeenCalled();
    expect(textarea).toHaveValue("hello\n");
  });

  it("does not send whitespace-only messages", async () => {
    const s = scenario();
    await s.user.type(
      screen.getByPlaceholderText("type a message..."),
      "   ",
    );
    expect(screen.getByText("send")).toBeDisabled();
  });
});
