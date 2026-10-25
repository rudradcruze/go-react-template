import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { pathToFileURL } from "node:url";

import openapiTS, { astToString } from "openapi-typescript";
import { describe, expect, it } from "vite-plus/test";

const specPath = resolve(import.meta.dirname, "../../../../backend/api/openapi.yaml");
const schemaPath = resolve(import.meta.dirname, "../../../src/api/schema.d.ts");

describe("the generated API types", () => {
  it("match the OpenAPI spec", async () => {
    const generated = astToString(await openapiTS(pathToFileURL(specPath)));

    // The file of the CLI has a comment header, and then the same text as the library.
    expect(
      readFileSync(schemaPath, "utf8").endsWith(generated),
      "The types are old. Run make gen.",
    ).toBe(true);
  });
});
