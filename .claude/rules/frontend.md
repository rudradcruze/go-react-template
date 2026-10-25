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
- Run tasks with the scripts in `frontend/package.json`.
- Use TypeScript in strict mode. Give each value a real type in place of `any`.

## Layout

```text
frontend/src/
├── api/                 typed API client and generated types
├── features/<context>/  screens and hooks for each bounded context
├── components/          shared UI components
└── routes/              React Router route definitions
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

- The Vite dev server uses port 5651 with `strictPort: true`.
- The Vite proxy sends `/api`, `/docs` and `/openapi.yaml` to `http://localhost:5653`.

## Accessibility

- Use semantic HTML elements. Give each form field a label.
- Make each action work with the keyboard.
