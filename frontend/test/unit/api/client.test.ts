import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vite-plus/test";

import { api } from "../../../src/api/client";
import { server } from "../../mocks/server";

describe("the API client", () => {
  it("sends requests to /api/v1 of the page origin, with the types of the spec", async () => {
    const user = {
      id: "0198f2a0-0000-7000-8000-000000000000",
      emailAddress: "ann@example.com",
      displayName: "Ann",
      createdAt: "2026-10-25T10:00:00Z",
    };
    server.use(http.post("/api/v1/users", () => HttpResponse.json(user, { status: 201 })));

    const { data, response } = await api.POST("/users", {
      body: { emailAddress: "ann@example.com", displayName: "Ann" },
    });

    expect(response.status).toBe(201);
    expect(data).toEqual(user);
  });
});
