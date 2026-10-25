import { useMutation } from "@tanstack/react-query";

import { api } from "../../api/client";
import { unwrap } from "../../api/errors";
import type { components } from "../../api/schema";

export type RegisterUserRequest = components["schemas"]["RegisterUserRequest"];

export function useRegisterUser() {
  return useMutation({
    mutationFn: async (body: RegisterUserRequest) => unwrap(await api.POST("/users", { body })),
  });
}
