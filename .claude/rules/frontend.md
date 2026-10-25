---
paths:
  - "frontend/src/**"
  - "frontend/index.html"
  - "frontend/package.json"
  - "frontend/vite.config.*"
---

# Frontend rules

## Toolchain

- The frontend uses React with TypeScript on the Vite+ toolchain. Vite+ builds, tests, lints and formats the code.
- Run tasks with the scripts in `frontend/package.json`. Vite+ reads all its settings from `frontend/vite.config.ts`.
- `pnpm-workspace.yaml` in the root of the repo holds the Vite+ catalog. Do not remove it: it replaces `vite` with the core of Vite+.
- Use TypeScript in strict mode. Give each value a real type in place of `any`.

## Layout

```text
frontend/src/
├── api/                 typed API client and generated types
├── features/<context>/  screens and hooks for each bounded context
├── components/          shared UI components
└── routes/              route definitions, and the pages that belong to no context
```

## API access

- `openapi-typescript` generates `src/api/schema.d.ts` from `backend/api/openapi.yaml`. Run `make gen` after a spec change. Do not edit the generated file.
- Send all API requests through the `openapi-fetch` client in `src/api/`. Components do not call `fetch` directly.
- Keep server data in TanStack Query. Start each query key with the context name, for example `["user", id]`.
- Give the response of the client to `unwrap` in `src/api/errors.ts`. It checks `response.ok`, not only the body, because a failed response can have no body. It throws an `ApiError` whose message is the `detail` of the problem, or a general message when the body has no text `detail`. Show that message to the user.

## Routes and styles

- Define routes with React Router in `src/routes/`.
- Make the query client with `createQueryClient` in `src/api/query-client.ts`. It sets the retry rule (`shouldRetry`) and a fresh time of 30 seconds for a query result.
- After a mutation, when the next page shows its result, put the result in the query cache with `setQueryData`, so the next page does not need a second request. Navigate with `replace` after a form, so the Back button does not return to the form.
- Each page sets the title of the tab with `usePageTitle`. A page that the person reaches after an action moves the focus to its heading.
- A page that shows data from a query keeps the data when a later refresh fails. It shows the error only when it has no data.
- Use Tailwind CSS v4 utility classes for styles.

## Dev server

- The Vite dev server uses port 5651 (`WEB_PORT`) with `strictPort: true`.
- The Vite proxy sends `/api`, `/docs` and `/openapi.yaml` to `http://localhost:5653` (`API_PORT`).
- `vite.config.ts` reads `WEB_PORT` and `API_PORT` from the `.env` file in the root of the repo, and from the environment.

## Accessibility

- Use semantic HTML elements. Give each form field a label.
- Make each action work with the keyboard. Do not use `disabled` on a button that the person has just used: a disabled button loses the focus. Use `aria-disabled` and ignore the second send.
- Keep the regions for the messages (`role="status"` and `role="alert"`) in the page all the time, and put the text in them. A screen reader announces the text that appears in a region that was there before.
