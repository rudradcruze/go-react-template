import { render, screen } from "@testing-library/react";
import { createMemoryRouter } from "react-router";
import { RouterProvider } from "react-router/dom";
import { describe, expect, it } from "vite-plus/test";

import { routes } from "../../../src/routes";

describe("the home page", () => {
  it("shows the name of the application", () => {
    render(<RouterProvider router={createMemoryRouter(routes, { initialEntries: ["/"] })} />);

    expect(
      screen.getByRole("heading", { level: 1, name: "go-react-template" }),
    ).toBeInTheDocument();
  });
});
