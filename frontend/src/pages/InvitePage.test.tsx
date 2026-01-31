import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi, beforeEach } from "vitest";
import InvitePage from "./InvitePage";
import type { Invite } from "../types";

const invite1: Invite = {
  id: 1,
  code: "abc-123-xyz",
  created_by: 1,
  used_by: null,
  created_at: "2024-01-01T00:00:00Z",
};

const invite2: Invite = {
  id: 2,
  code: "def-456-uvw",
  created_by: 1,
  used_by: 5,
  created_at: "2024-01-02T00:00:00Z",
};

const mockUseInvites = {
  invites: [invite1, invite2],
  loading: false,
  createInvite: vi.fn(),
  refresh: vi.fn(),
};

vi.mock("../hooks/useInvites", () => ({
  useInvites: () => mockUseInvites,
}));

function scenario(overrides?: Partial<typeof mockUseInvites>) {
  Object.assign(mockUseInvites, overrides);
  const user = userEvent.setup();
  render(<InvitePage />);
  return { user };
}

beforeEach(() => {
  vi.resetAllMocks();
  Object.assign(mockUseInvites, {
    invites: [invite1, invite2],
    loading: false,
    createInvite: vi.fn(),
    refresh: vi.fn(),
  });
});

describe("InvitePage", () => {
  it("renders invite list", () => {
    scenario();
    expect(screen.getByText("abc-123-xyz")).toBeInTheDocument();
    expect(screen.getByText("def-456-uvw")).toBeInTheDocument();
  });

  it("shows invite count", () => {
    scenario();
    expect(screen.getByText("2 invites generated")).toBeInTheDocument();
  });

  it("shows empty state when no invites", () => {
    scenario({ invites: [] });
    expect(screen.getByText(/no invites yet/)).toBeInTheDocument();
  });

  it("shows available status for unused invite", () => {
    scenario();
    expect(screen.getByText("available")).toBeInTheDocument();
  });

  it("shows used status for used invite", () => {
    scenario();
    expect(screen.getByText("used")).toBeInTheDocument();
  });

  it("shows copy button for unused invites only", () => {
    scenario();
    const copyButtons = screen.getAllByText("copy");
    expect(copyButtons).toHaveLength(1);
  });

  it("calls createInvite when generate button is clicked", async () => {
    mockUseInvites.createInvite.mockResolvedValue({
      id: 3,
      code: "new-code",
      created_by: 1,
      used_by: null,
      created_at: "2024-01-03T00:00:00Z",
    });
    const { user } = scenario();
    await user.click(screen.getByText("+ Generate Invite"));
    expect(mockUseInvites.createInvite).toHaveBeenCalledOnce();
  });

  it("shows copied state after clicking copy", async () => {
    // Stub clipboard in jsdom
    Object.defineProperty(navigator, "clipboard", {
      value: { writeText: vi.fn().mockResolvedValue(undefined) },
      configurable: true,
    });

    const { user } = scenario();
    await user.click(screen.getByText("copy"));
    expect(await screen.findByText("copied!")).toBeInTheDocument();
  });
});
