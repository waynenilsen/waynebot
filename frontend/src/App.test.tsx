import { render } from "@testing-library/react";
import { expect, test } from "vitest";
import App from "./App";

test("App mounts without crashing", () => {
  const { container } = render(<App />);
  expect(container).toBeInTheDocument();
});
