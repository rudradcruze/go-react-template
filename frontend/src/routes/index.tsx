import type { RouteObject } from "react-router";

import { RegisterPage } from "../features/user/register-page";
import { UserPage } from "../features/user/user-page";
import { HomePage } from "./home-page";

export const routes: RouteObject[] = [
  { path: "/", element: <HomePage /> },
  { path: "/register", element: <RegisterPage /> },
  { path: "/users/:id", element: <UserPage /> },
];
