import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vite-plus/test";

import { api } from "../../../src/api/client";
import { server } from "../../mocks/server";

describe("the API client", () => {
  it("sends requests to /api/v1 of the page origin", async () => {
    server.use(http.get("/api/v1/ping", () => HttpResponse.json({ ok: true })));

    // The spec has no paths yet, so TypeScript accepts any path. Use a typed path when the spec has one.
    const { data } = await api.GET("/ping");

    expect(data).toEqual({ ok: true });
  });
});
