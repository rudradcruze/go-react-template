import type { components } from "../../src/api/schema";

export type Problem = components["schemas"]["Problem"];

export function problem(status: number, overrides: Partial<Problem> = {}): Problem {
  return {
    type: "about:blank",
    title: "Problem",
    status,
    detail: "The request is not valid. Correct the request.",
    ...overrides,
  };
}
