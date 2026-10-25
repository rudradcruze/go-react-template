import type { RequestHandler } from "msw";

// These are the default handlers of all tests. A test adds its own handlers with server.use.
export const handlers: RequestHandler[] = [];
