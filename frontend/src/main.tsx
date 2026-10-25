import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { createBrowserRouter } from "react-router";
import { RouterProvider } from "react-router/dom";

import { routes } from "./routes";
import "./styles.css";

const container = document.getElementById("root");
if (!container) {
  throw new Error(
    "The page has no element with the ID root. Add an element with the ID root to index.html.",
  );
}

const queryClient = new QueryClient();
const router = createBrowserRouter(routes);

createRoot(container).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </StrictMode>,
);
