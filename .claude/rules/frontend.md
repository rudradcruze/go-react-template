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
- Show the `detail` text of a problem response to the user when it helps the user correct the problem.

## Routes and styles

- Define routes with React Router in `src/routes/`.
- Use Tailwind CSS v4 utility classes for styles.

## Dev server

- The Vite dev server uses port 5651 (`WEB_PORT`) with `strictPort: true`.
- The Vite proxy sends `/api`, `/docs` and `/openapi.yaml` to `http://localhost:5653` (`API_PORT`).
- `vite.config.ts` reads `WEB_PORT` and `API_PORT` from the `.env` file in the root of the repo, and from the environment.

## Accessibility

- Use semantic HTML elements. Give each form field a label.
- Make each action work with the keyboard.
