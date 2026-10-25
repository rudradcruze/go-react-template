import { useQuery } from "@tanstack/react-query";

import { api } from "../../api/client";
import { unwrap } from "../../api/errors";

// userQueryKey is the key of a User in the cache of the queries. It starts with the name of the context.
export const userQueryKey = (id: string) => ["user", id] as const;

export function useUser(id: string) {
  return useQuery({
    queryKey: userQueryKey(id),
    queryFn: async () => unwrap(await api.GET("/users/{id}", { params: { path: { id } } })),
  });
}
