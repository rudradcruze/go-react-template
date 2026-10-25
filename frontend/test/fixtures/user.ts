import type { components } from "../../src/api/schema";

export type User = components["schemas"]["User"];
export type RegisterUserRequest = components["schemas"]["RegisterUserRequest"];

// The ID of a User that does not exist.
export const unknownUserId = "0198f2a0-0000-7000-8000-000000000000";

let counter = 0;

// A builder gives a valid object with default values, and each call gives unique values for the unique fields.
export function user(overrides: Partial<User> = {}): User {
  counter += 1;
  return {
    id: `0198f2a0-0000-7000-8000-${String(counter).padStart(12, "0")}`,
    emailAddress: `user${counter}@example.com`,
    displayName: `User ${counter}`,
    createdAt: "2026-10-25T10:00:00Z",
    ...overrides,
  };
}

export function registerUserRequest(
  overrides: Partial<RegisterUserRequest> = {},
): RegisterUserRequest {
  counter += 1;
  return {
    emailAddress: `user${counter}@example.com`,
    displayName: `User ${counter}`,
    ...overrides,
  };
}
