import createClient from "openapi-fetch";

import type { paths } from "./schema";

// The paths of the spec are relative to its server URL, /api/v1.
// The base URL is absolute, because a request with a relative URL fails outside a browser, for example in a test.
// The API is on the origin of the page: the Vite proxy and nginx send /api to the API.
export const api = createClient<paths>({ baseUrl: `${window.location.origin}/api/v1` });
