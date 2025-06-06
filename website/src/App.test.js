import { render, screen } from "@testing-library/react";
import App from "./App";

test("renders Crawler System Testing header", () => {
  render(<App />);
  const headerElement = screen.getByRole("heading", {
    name: /Crawler System Testing 1/i,
  });
  expect(headerElement).toBeInTheDocument();
});
