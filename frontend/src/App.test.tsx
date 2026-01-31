import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi, beforeEach } from "vitest";
import App from "./App";
import type { AuthResponse, User } from "./types";

const alice: User = {
  id: 1,
  username: "alice",
  created_at: "2024-01-01T00:00:00Z",
};
const authResp: AuthResponse = { token: "tok_abc", user: alice };

vi.mock("./api", () => ({
  login: vi.fn(),
  register: vi.fn(),
  logout: vi.fn(),
  getMe: vi.fn(),
}));

vi.mock("./utils/token", () => ({
  getToken: vi.fn(),
  setToken: vi.fn(),
  clearToken: vi.fn(),
}));

import * as api from "./api";
import * as tokenUtils from "./utils/token";

const mockApi = vi.mocked(api);
const mockToken = vi.mocked(tokenUtils);

beforeEach(() => {
  vi.resetAllMocks();
  mockToken.getToken.mockReturnValue(null);
});

describe("App", () => {
  it("shows loading state initially when token exists", () => {
    mockToken.getToken.mockReturnValue("tok_abc");
    mockApi.getMe.mockReturnValue(new Promise(() => {})); // never resolves
    render(<App />);
    expect(screen.getByText("loading...")).toBeInTheDocument();
  });

  it("shows login page when no user", async () => {
    render(<App />);
    await waitFor(() =>
      expect(screen.getByText("Sign in to your workspace")).toBeInTheDocument(),
    );
  });

  it("shows welcome message after login", async () => {
    mockApi.login.mockResolvedValue(authResp);
    const user = userEvent.setup();

    render(<App />);
    await waitFor(() =>
      expect(screen.getByLabelText("Username")).toBeInTheDocument(),
    );

    await user.type(screen.getByLabelText("Username"), "alice");
    await user.type(screen.getByLabelText("Password"), "password123");
    await user.click(
      document.querySelector("button[type='submit']") as HTMLButtonElement,
    );

    await waitFor(() => expect(screen.getByText("alice")).toBeInTheDocument());
    expect(screen.getByText(/welcome,/i)).toBeInTheDocument();
  });

  it("restores session from token and shows welcome", async () => {
    mockToken.getToken.mockReturnValue("tok_abc");
    mockApi.getMe.mockResolvedValue(alice);

    render(<App />);
    await waitFor(() => expect(screen.getByText("alice")).toBeInTheDocument());
  });

  it("shows login page after logout", async () => {
    mockToken.getToken.mockReturnValue("tok_abc");
    mockApi.getMe.mockResolvedValue(alice);
    mockApi.logout.mockResolvedValue(undefined);
    const user = userEvent.setup();

    render(<App />);
    await waitFor(() => expect(screen.getByText("alice")).toBeInTheDocument());

    await user.click(screen.getByText("Sign out"));

    await waitFor(() =>
      expect(screen.getByText("Sign in to your workspace")).toBeInTheDocument(),
    );
  });
});
