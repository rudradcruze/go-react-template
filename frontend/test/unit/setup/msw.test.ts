import { http, HttpResponse } from "msw";
import { describe, expect, it } from "vite-plus/test";

import { server } from "../../mocks/server";
import { unhandledRequests } from "../../mocks/unhandled";

const api = "http://localhost:5653/api/v1";

describe("the mock server", () => {
  it("records a request that has no handler, and does not send it to the network", async () => {
    const response = await fetch(`${api}/without-handler`);

    // A connection error would mean that the request left the mock server.
    expect(response.status).toBe(500);
    expect(await response.text()).toContain("The request has no handler");
    // The test takes the record, so the afterEach hook in test/setup.ts does not fail this test.
    const recorded = unhandledRequests.splice(0);
    expect(recorded).toHaveLength(1);
    expect(recorded[0]).toContain("/api/v1/without-handler");
  });

  it("answers a request that has a handler", async () => {
    server.use(http.get(`${api}/ping`, () => HttpResponse.json({ ok: true })));

    const response = await fetch(`${api}/ping`);

    expect(await response.json()).toEqual({ ok: true });
  });
});
