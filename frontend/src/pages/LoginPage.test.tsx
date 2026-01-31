import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import LoginPage from "./LoginPage";

function scenario(overrides?: {
  onLogin?: (u: string, p: string) => Promise<void>;
  onRegister?: (u: string, p: string, c?: string) => Promise<void>;
}) {
  const onLogin = overrides?.onLogin ?? vi.fn(() => Promise.resolve());
  const onRegister = overrides?.onRegister ?? vi.fn(() => Promise.resolve());
  const user = userEvent.setup();

  render(<LoginPage onLogin={onLogin} onRegister={onRegister} />);

  return {
    user,
    onLogin: onLogin as ReturnType<typeof vi.fn>,
    onRegister: onRegister as ReturnType<typeof vi.fn>,
    username: () => screen.getByLabelText("Username"),
    password: () => screen.getByLabelText("Password"),
    submit: () =>
      document.querySelector("button[type='submit']") as HTMLButtonElement,
    signInTab: () => screen.getByTestId("tab-login"),
    registerTab: () => screen.getByTestId("tab-register"),
  };
}

describe("LoginPage", () => {
  it("renders login mode by default", () => {
    scenario();
    expect(screen.getByText("Sign in to your workspace")).toBeInTheDocument();
    expect(screen.getByLabelText("Username")).toBeInTheDocument();
    expect(screen.getByLabelText("Password")).toBeInTheDocument();
    expect(screen.queryByLabelText(/invite code/i)).not.toBeInTheDocument();
  });

  it("toggles to register mode", async () => {
    const s = scenario();
    await s.user.click(s.registerTab());
    expect(screen.getByText("Create your account")).toBeInTheDocument();
    expect(screen.getByLabelText(/invite code/i)).toBeInTheDocument();
  });

  it("toggles back to login mode", async () => {
    const s = scenario();
    await s.user.click(s.registerTab());
    await s.user.click(s.signInTab());
    expect(screen.getByText("Sign in to your workspace")).toBeInTheDocument();
    expect(screen.queryByLabelText(/invite code/i)).not.toBeInTheDocument();
  });

  it("calls onLogin with username and password", async () => {
    const s = scenario();
    await s.user.type(s.username(), "alice");
    await s.user.type(s.password(), "password123");
    await s.user.click(s.submit());

    expect(s.onLogin).toHaveBeenCalledWith("alice", "password123");
  });

  it("calls onRegister with username, password, and invite code", async () => {
    const s = scenario();
    await s.user.click(s.registerTab());
    await s.user.type(s.username(), "bob");
    await s.user.type(s.password(), "secret1234");
    await s.user.type(screen.getByLabelText(/invite code/i), "inv_abc");
    await s.user.click(s.submit());

    expect(s.onRegister).toHaveBeenCalledWith("bob", "secret1234", "inv_abc");
  });

  it("calls onRegister without invite code when empty", async () => {
    const s = scenario();
    await s.user.click(s.registerTab());
    await s.user.type(s.username(), "bob");
    await s.user.type(s.password(), "secret1234");
    await s.user.click(s.submit());

    expect(s.onRegister).toHaveBeenCalledWith("bob", "secret1234", undefined);
  });

  it("displays error on login failure", async () => {
    const s = scenario({
      onLogin: vi.fn(() => Promise.reject(new Error("Invalid credentials"))),
    });
    await s.user.type(s.username(), "alice");
    await s.user.type(s.password(), "wrongpass1");
    await s.user.click(s.submit());

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Invalid credentials",
    );
  });

  it("displays error on register failure", async () => {
    const s = scenario({
      onRegister: vi.fn(() => Promise.reject(new Error("Username taken"))),
    });
    await s.user.click(s.registerTab());
    await s.user.type(s.username(), "alice");
    await s.user.type(s.password(), "password123");
    await s.user.click(s.submit());

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Username taken",
    );
  });

  it("shows username validation hint for invalid input", async () => {
    const s = scenario();
    await s.user.type(s.username(), "no spaces!");
    expect(
      screen.getByText(/letters, numbers, and underscores only/i),
    ).toBeInTheDocument();
  });

  it("shows password validation hint when too short", async () => {
    const s = scenario();
    await s.user.type(s.password(), "short");
    expect(screen.getByText(/must be 8-128 characters/i)).toBeInTheDocument();
  });

  it("clears error when switching modes", async () => {
    const s = scenario({
      onLogin: vi.fn(() => Promise.reject(new Error("Bad"))),
    });
    await s.user.type(s.username(), "alice");
    await s.user.type(s.password(), "password123");
    await s.user.click(s.submit());

    expect(await screen.findByRole("alert")).toBeInTheDocument();

    await s.user.click(s.registerTab());
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});
