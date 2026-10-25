import { useEffect, useRef } from "react";
import { useParams } from "react-router";

import { messageOf } from "../../api/errors";
import { usePageTitle } from "../../components/use-page-title";
import { useUser } from "./use-user";

export function UserPage() {
  usePageTitle("User");
  const { id = "" } = useParams();
  const user = useUser(id);

  // The person comes here from another page, so move the focus to the heading. A screen reader then reads it.
  const heading = useRef<HTMLHeadingElement>(null);
  useEffect(() => heading.current?.focus(), []);

  return (
    <main className="mx-auto max-w-md p-8">
      <h1 ref={heading} tabIndex={-1} className="text-3xl font-bold focus:outline-none">
        User
      </h1>
      {/* The 2 regions are always in the page, so a screen reader announces the text that appears in them. */}
      <div className="mt-4 empty:mt-0" role="status">
        {user.isPending && "The application loads the User."}
      </div>
      {/* A failed refresh keeps the User that the page already shows. Only a page without a User shows the error. */}
      <div className="mt-4 text-red-800 empty:mt-0" role="alert">
        {user.isError && user.data === undefined && messageOf(user.error)}
      </div>
      {user.data !== undefined && (
        <dl className="mt-6 grid grid-cols-[auto_1fr] gap-x-4 gap-y-2">
          <dt className="font-bold">Email address</dt>
          <dd>{user.data.emailAddress}</dd>
          <dt className="font-bold">Display name</dt>
          <dd>{user.data.displayName}</dd>
        </dl>
      )}
    </main>
  );
}
