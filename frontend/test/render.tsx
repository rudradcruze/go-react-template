import { QueryClientProvider } from "@tanstack/react-query";
import { render } from "@testing-library/react";
import { createMemoryRouter } from "react-router";
import { RouterProvider } from "react-router/dom";

import { createQueryClient } from "../src/api/query-client";
import { routes } from "../src/routes";

// Renders the application at a path, with a new query client that does not try again after an error.
// It gives the router and the query client, so a test can check the path of the page and refresh a query.
export function renderApp(path: string) {
  const queryClient = createQueryClient(false);
  const router = createMemoryRouter(routes, { initialEntries: [path] });
  const view = render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
  return { ...view, router, queryClient };
}
