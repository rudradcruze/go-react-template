import { act, screen, waitFor } from "@testing-library/react";
import { delay, http, HttpResponse } from "msw";
import { describe, expect, it } from "vite-plus/test";

import { problem, type Problem } from "../../fixtures/problem";
import { unknownUserId, user, type User } from "../../fixtures/user";
import { userUrl } from "../../mocks/urls";
import { server } from "../../mocks/server";
import { userQueryKey } from "../../../src/features/user/use-user";
import { renderApp } from "../../render";

type UserParams = { id: string };

describe("the User page", () => {
  it("shows the email address and the display name of the User", async () => {
    const ann = user({ emailAddress: "ann@example.com", displayName: "Ann" });
    server.use(http.get<UserParams, never, User>(userUrl, () => HttpResponse.json(ann)));
    renderApp(`/users/${ann.id}`);

    expect(await screen.findByText("ann@example.com")).toBeInTheDocument();
    expect(screen.getByText("Email address").nextElementSibling).toHaveTextContent(
      "ann@example.com",
    );
    expect(screen.getByText("Display name").nextElementSibling).toHaveTextContent("Ann");
  });

  it("asks the API for the User of the URL", async () => {
    const requested: string[] = [];
    const ann = user();
    server.use(
      http.get<UserParams, never, User>(userUrl, ({ params }) => {
        requested.push(params.id);
        return HttpResponse.json(ann);
      }),
    );
    renderApp(`/users/${ann.id}`);

    await screen.findByText(ann.emailAddress);

    expect(requested).toEqual([ann.id]);
  });

  it("has a heading, a page title and the focus on the heading", () => {
    server.use(http.get<UserParams, never, User>(userUrl, () => HttpResponse.json(user())));
    renderApp(`/users/${unknownUserId}`);

    const heading = screen.getByRole("heading", { level: 1, name: "User" });
    expect(heading).toHaveFocus();
    expect(document.title).toBe("User - go-react-template");
  });

  it("shows a loading message until the User arrives", async () => {
    const ann = user();
    server.use(
      http.get<UserParams, never, User>(userUrl, async () => {
        await delay(100);
        return HttpResponse.json(ann);
      }),
    );
    renderApp(`/users/${ann.id}`);

    expect(screen.getByRole("status")).toHaveTextContent("The application loads the User.");
    expect(screen.queryByText(ann.emailAddress)).not.toBeInTheDocument();
    await screen.findByText(ann.emailAddress);
  });

  it("removes the loading message when the User arrives", async () => {
    const ann = user();
    server.use(http.get<UserParams, never, User>(userUrl, () => HttpResponse.json(ann)));
    renderApp(`/users/${ann.id}`);

    await screen.findByText(ann.emailAddress);

    expect(screen.getByRole("status")).toBeEmptyDOMElement();
  });

  it("shows the message of the API when no User has the ID", async () => {
    const detail = "The User does not exist. Check the ID.";
    server.use(
      http.get<UserParams, never, Problem>(userUrl, () =>
        HttpResponse.json(problem(404, { detail }), { status: 404 }),
      ),
    );
    renderApp(`/users/${unknownUserId}`);

    await screen.findByText(detail);
    expect(screen.getByRole("alert")).toHaveTextContent(detail);
    expect(screen.queryByText("Email address")).not.toBeInTheDocument();
    expect(screen.getByRole("status")).toBeEmptyDOMElement();
  });

  it("shows a general message for a 404 that has no problem detail", async () => {
    server.use(http.get(userUrl, () => HttpResponse.html("<h1>Not found</h1>", { status: 404 })));
    renderApp(`/users/${unknownUserId}`);

    await screen.findByText("The server cannot find the data. Check the URL of the page.");
  });

  it.each([
    [400, "The ID is not valid. Use the ID of a User, in the format of a UUID."],
    [500, "The server cannot complete the request. Try again in 1 minute."],
  ])("shows the detail of a problem with the status %i", async (status, detail) => {
    server.use(
      http.get<UserParams, never, Problem>(userUrl, () =>
        HttpResponse.json(problem(status, { detail }), { status }),
      ),
    );
    renderApp("/users/abc");

    await screen.findByText(detail);
    expect(screen.getByRole("alert")).toHaveTextContent(detail);
  });

  it("shows a message when the server does not answer", async () => {
    server.use(http.get(userUrl, () => HttpResponse.error()));
    renderApp(`/users/${unknownUserId}`);

    await screen.findByText(/The application cannot reach the server\./);
  });

  it("keeps the User on the screen when a later refresh fails", async () => {
    const ann = user();
    let fail = false;
    server.use(
      http.get<UserParams, never, User | Problem>(userUrl, () =>
        fail ? HttpResponse.json(problem(500), { status: 500 }) : HttpResponse.json(ann),
      ),
    );
    const { queryClient } = renderApp(`/users/${ann.id}`);
    await screen.findByText(ann.emailAddress);

    fail = true;
    await queryClient.invalidateQueries();
    // The query must have failed, and React must have shown the new state, before the test looks at the page.
    await waitFor(() =>
      expect(queryClient.getQueryState(userQueryKey(ann.id))?.status).toBe("error"),
    );
    await act(async () => {});

    expect(screen.getByText(ann.emailAddress)).toBeInTheDocument();
    expect(screen.getByRole("alert")).toBeEmptyDOMElement();
  });
});
