import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi, beforeEach } from "vitest";
import PersonaPage from "./PersonaPage";
import type { Persona } from "../types";

const alice: Persona = {
  id: 1,
  name: "alice",
  system_prompt: "You are Alice, a helpful assistant",
  model: "gpt-4",
  tools_enabled: ["web_search"],
  temperature: 0.7,
  max_tokens: 4096,
  cooldown_secs: 5,
  max_tokens_per_hour: 100000,
  created_at: "2024-01-01T00:00:00Z",
};

const bob: Persona = {
  id: 2,
  name: "bob",
  system_prompt: "You are Bob",
  model: "claude-3-opus",
  tools_enabled: [],
  temperature: 1.0,
  max_tokens: 8192,
  cooldown_secs: 0,
  max_tokens_per_hour: 50000,
  created_at: "2024-01-01T00:00:00Z",
};

const mockUsePersonas = {
  personas: [alice, bob],
  loading: false,
  createPersona: vi.fn(),
  updatePersona: vi.fn(),
  deletePersona: vi.fn(),
  refresh: vi.fn(),
};

vi.mock("../hooks/usePersonas", () => ({
  usePersonas: () => mockUsePersonas,
}));

function scenario(overrides?: Partial<typeof mockUsePersonas>) {
  Object.assign(mockUsePersonas, overrides);
  const user = userEvent.setup();
  render(<PersonaPage />);
  return { user };
}

beforeEach(() => {
  vi.resetAllMocks();
  Object.assign(mockUsePersonas, {
    personas: [alice, bob],
    loading: false,
    createPersona: vi.fn(),
    updatePersona: vi.fn(),
    deletePersona: vi.fn(),
    refresh: vi.fn(),
  });
});

describe("PersonaPage", () => {
  it("renders persona list", () => {
    scenario();
    expect(screen.getByText("alice")).toBeInTheDocument();
    expect(screen.getByText("bob")).toBeInTheDocument();
  });

  it("shows persona count", () => {
    scenario();
    expect(screen.getByText("2 personas configured")).toBeInTheDocument();
  });

  it("shows empty state when no personas", () => {
    scenario({ personas: [] });
    expect(screen.getByText(/no personas yet/)).toBeInTheDocument();
  });

  it("shows loading state", () => {
    scenario({ personas: [], loading: true });
    expect(screen.getByText("loading...")).toBeInTheDocument();
  });

  it("shows model name for each persona", () => {
    scenario();
    expect(screen.getByText("gpt-4")).toBeInTheDocument();
    expect(screen.getByText("claude-3-opus")).toBeInTheDocument();
  });

  it("opens create form when clicking new persona", async () => {
    const { user } = scenario();
    await user.click(screen.getByText("+ New Persona"));
    expect(screen.getByText("New Persona")).toBeInTheDocument();
    expect(screen.getByLabelText(/name/i)).toBeInTheDocument();
  });

  it("shows delete confirmation", async () => {
    const { user } = scenario();
    const deleteButtons = screen.getAllByText("delete");
    await user.click(deleteButtons[0]);
    expect(screen.getByText("confirm")).toBeInTheDocument();
    expect(screen.getByText("cancel")).toBeInTheDocument();
  });

  it("calls deletePersona on confirm", async () => {
    mockUsePersonas.deletePersona.mockResolvedValue(undefined);
    const { user } = scenario();
    const deleteButtons = screen.getAllByText("delete");
    await user.click(deleteButtons[0]);
    await user.click(screen.getByText("confirm"));
    expect(mockUsePersonas.deletePersona).toHaveBeenCalledWith(1);
  });

  it("cancels delete confirmation", async () => {
    const { user } = scenario();
    const deleteButtons = screen.getAllByText("delete");
    await user.click(deleteButtons[0]);
    await user.click(screen.getByText("cancel"));
    expect(screen.queryByText("confirm")).not.toBeInTheDocument();
  });

  it("opens edit form when clicking edit", async () => {
    const { user } = scenario();
    const editButtons = screen.getAllByText("edit");
    await user.click(editButtons[0]);
    expect(screen.getByText("Edit Persona")).toBeInTheDocument();
    expect(screen.getByDisplayValue("alice")).toBeInTheDocument();
  });
});
