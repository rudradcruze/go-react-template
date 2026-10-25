import { expect, test, type APIRequestContext } from "@playwright/test";

import { newPerson } from "./support";

// These tests check what nginx does between the browser and the API.

// postRegistration sends a Registration with a display name of the given length. The body is about
// as large as the display name.
function postRegistration(request: APIRequestContext, displayNameLength: number) {
  const body = JSON.stringify({
    emailAddress: newPerson().emailAddress,
    displayName: "a".repeat(displayNameLength),
  });
  return request.post("/api/v1/users", {
    headers: { "content-type": "application/json" },
    data: body,
  });
}

test("nginx answers its own health check", async ({ request }) => {
  const response = await request.get("/healthz");

  expect(response.status()).toBe(200);
  expect(await response.text()).toBe("ok\n");
});

test("nginx sets the security headers on a page of the frontend", async ({ request }) => {
  const headers = (await request.get("/register")).headers();

  expect(headers["x-content-type-options"]).toBe("nosniff");
  expect(headers["x-frame-options"]).toBe("DENY");
  expect(headers["referrer-policy"]).toBe("no-referrer");
});

test("nginx sets a CSP on a page of the frontend, and not on the API docs, which load files from other hosts", async ({
  request,
}) => {
  const page = await request.get("/register");
  const docs = await request.get("/docs");

  expect(page.headers()["content-security-policy"]).toContain("default-src 'self'");
  expect(docs.headers()["content-security-policy"]).toBeUndefined();
});

for (const path of ["/api", "/api/", "/api/v1/unknown", "/docs", "/docs/", "/openapi.yaml/"]) {
  test(`nginx sends ${path} to the API, which answers with a problem detail`, async ({
    request,
  }) => {
    const response = await request.get(path);

    expect(response.headers()["content-type"]).toContain("application/problem+json");
  });
}

test("nginx does not give the page of the frontend for a file that does not exist", async ({
  request,
}) => {
  const response = await request.get("/assets/does-not-exist.js");

  expect(response.status()).toBe(404);
});

// The API limit is 1 MiB (1,048,576 bytes), and the nginx limit is 2 MiB. A display name of 1,500,000 characters is
// between the 2 limits, and one of 2,600,000 characters is above both.
test("the API answers a body of 1.5 MB with its own 413 problem detail", async ({ request }) => {
  const response = await postRegistration(request, 1_500_000);

  expect(response.status()).toBe(413);
  expect(response.headers()["content-type"]).toContain("application/problem+json");
  expect((await response.json()).detail).toContain("larger than 1 MiB");
});

test("nginx answers a body of 2.6 MB with a 413 problem detail of the same kind", async ({
  request,
}) => {
  const response = await postRegistration(request, 2_600_000);

  expect(response.status()).toBe(413);
  expect(response.headers()["content-type"]).toContain("application/problem+json");
  expect((await response.json()).detail).toContain("larger than 2 MiB");
});
