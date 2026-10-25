import { useState, type FormEvent } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router";

import { messageOf } from "../../api/errors";
import { usePageTitle } from "../../components/use-page-title";
import { useRegisterUser, type RegisterUserRequest } from "./use-register-user";
import { userQueryKey } from "./use-user";

const fieldClass = "mt-1 w-full rounded border border-gray-500 px-3 py-2";

export function RegisterPage() {
  usePageTitle("Registration");
  const [form, setForm] = useState<RegisterUserRequest>({ emailAddress: "", displayName: "" });
  const registration = useRegisterUser();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (registration.isPending) {
      return;
    }
    registration.mutate(form, {
      onSuccess: (user) => {
        // The response has the User, so the User page does not depend on a second request.
        queryClient.setQueryData(userQueryKey(user.id), user);
        // Replace the page in the history, so the Back button does not return to the form: a second send gives 409.
        void navigate(`/users/${user.id}`, { replace: true });
      },
    });
  }

  return (
    <main className="mx-auto max-w-md p-8">
      <h1 className="text-3xl font-bold">Registration</h1>
      {/* The server checks the values and gives the messages, so the browser does not check them. */}
      <form className="mt-6 flex flex-col gap-4" onSubmit={submit} noValidate>
        <div>
          <label htmlFor="emailAddress">Email address</label>
          <input
            id="emailAddress"
            className={fieldClass}
            type="email"
            autoComplete="email"
            value={form.emailAddress}
            onChange={(event) => setForm({ ...form, emailAddress: event.target.value })}
          />
        </div>
        <div>
          <label htmlFor="displayName">Display name</label>
          {/* A display name is not the full name of the person, so the browser must not fill it with the full name. */}
          <input
            id="displayName"
            className={fieldClass}
            type="text"
            autoComplete="nickname"
            value={form.displayName}
            onChange={(event) => setForm({ ...form, displayName: event.target.value })}
          />
        </div>
        {/* The button uses aria-disabled, not disabled: a disabled button loses the keyboard focus. submit ignores a second send. */}
        <button
          className="rounded bg-blue-800 px-4 py-2 text-white aria-disabled:opacity-50"
          type="submit"
          aria-disabled={registration.isPending}
        >
          Register
        </button>
      </form>
      {/* The region is always in the page. A screen reader announces the text that appears in a region that was there before. */}
      <div className="mt-4 text-red-800 empty:mt-0" role="alert">
        {registration.isError && messageOf(registration.error)}
      </div>
    </main>
  );
}
