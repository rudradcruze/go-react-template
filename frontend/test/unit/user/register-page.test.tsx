import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { delay, http, HttpResponse } from "msw";
import { describe, expect, it } from "vite-plus/test";

import { problem, type Problem } from "../../fixtures/problem";
import {
  registerUserRequest,
  user,
  type RegisterUserRequest,
  type User,
} from "../../fixtures/user";
import { server } from "../../mocks/server";
import { registerUrl, userUrl } from "../../mocks/urls";
import { renderApp } from "../../render";

const serverProblem = "The server cannot complete the request. Try again in 1 minute.";
const rejectedRequest = "The server did not accept the request. Check the values. Then try again.";

async function fillAndSubmit(request: RegisterUserRequest) {
  const person = userEvent.setup();
  await person.type(screen.getByLabelText("Email address"), request.emailAddress);
  await person.type(screen.getByLabelText("Display name"), request.displayName);
  await person.click(screen.getByRole("button", { name: "Register" }));
}

// registrationHandlers answer with a User, and it saves the body of the request.
// The next page, the User page, reads the same User.
function registrationHandlers(received: RegisterUserRequest[]) {
  let registered: User | undefined;
  return [
    http.post<never, RegisterUserRequest, User>(registerUrl, async ({ request }) => {
      const body = await request.json();
      received.push(body);
      registered = user(body);
      return HttpResponse.json(registered, { status: 201 });
    }),
    http.get<{ id: string }, never, User>(userUrl, () => HttpResponse.json(registered)),
  ];
}

describe("the Registration page", () => {
  it("has a label for each field and a button", () => {
    renderApp("/register");

    expect(screen.getByRole("heading", { level: 1, name: "Registration" })).toBeInTheDocument();
    expect(screen.getByLabelText("Email address")).toBeInTheDocument();
    expect(screen.getByLabelText("Display name")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Register" })).toBeInTheDocument();
  });

  it("has the region of the error message before it has a message, so that a screen reader announces the first message", () => {
    renderApp("/register");

    expect(screen.getByRole("alert")).toBeEmptyDOMElement();
  });

  it("sends the email address and the display name that the person entered", async () => {
    const received: RegisterUserRequest[] = [];
    server.use(...registrationHandlers(received));
    renderApp("/register");
    const request = registerUserRequest({ emailAddress: "ann@example.com", displayName: "Ann" });

    await fillAndSubmit(request);

    await screen.findByRole("heading", { level: 1, name: "User" });
    expect(received).toEqual([request]);
  });

  it("opens the User page of the new User after a Registration", async () => {
    server.use(...registrationHandlers([]));
    const { router } = renderApp("/register");

    await fillAndSubmit(
      registerUserRequest({ emailAddress: "ann@example.com", displayName: "Ann" }),
    );

    expect(await screen.findByText("ann@example.com")).toBeInTheDocument();
    expect(screen.getByText("Ann")).toBeInTheDocument();
    expect(router.state.location.pathname).toMatch(/^\/users\/[0-9a-f-]{36}$/);
  });

  it("opens the User page with the User of the Registration, without a second request", async () => {
    // The test has no handler for the User page. A request for the User would fail the test.
    const ann = user({ emailAddress: "ann@example.com", displayName: "Ann" });
    server.use(
      http.post<never, RegisterUserRequest, User>(registerUrl, () =>
        HttpResponse.json(ann, { status: 201 }),
      ),
    );
    renderApp("/register");

    await fillAndSubmit(registerUserRequest());

    expect(await screen.findByText("ann@example.com")).toBeInTheDocument();
    expect(screen.getByText("Ann")).toBeInTheDocument();
    expect(screen.getByRole("alert")).toBeEmptyDOMElement();
  });

  it("replaces the Registration page in the history, so the Back button does not return to an empty form", async () => {
    server.use(...registrationHandlers([]));
    const { router } = renderApp("/register");

    await fillAndSubmit(registerUserRequest());

    await screen.findByRole("heading", { level: 1, name: "User" });
    expect(router.state.historyAction).toBe("REPLACE");
  });

  it("has a page title", () => {
    renderApp("/register");

    expect(document.title).toBe("Registration - go-react-template");
  });

  it("stays on the Registration page after a failure", async () => {
    server.use(
      http.post<never, RegisterUserRequest, Problem>(registerUrl, () =>
        HttpResponse.json(problem(409, { detail: "The email address is already registered." }), {
          status: 409,
        }),
      ),
    );
    const { router } = renderApp("/register");

    await fillAndSubmit(registerUserRequest());

    await screen.findByText("The email address is already registered.");
    expect(router.state.location.pathname).toBe("/register");
  });

  it.each([
    [409, "The email address is already registered. Use a different email address."],
    [422, "The email address is not valid. Enter an address in the format name@example.com."],
    [500, serverProblem],
  ])("shows the detail of a problem with the status %i", async (status, detail) => {
    server.use(
      http.post(registerUrl, () => HttpResponse.json(problem(status, { detail }), { status })),
    );
    renderApp("/register");

    await fillAndSubmit(registerUserRequest());

    expect(await screen.findByText(detail)).toBeInTheDocument();
    expect(screen.getByRole("alert")).toHaveTextContent(detail);
  });

  it.each([
    ["an empty body from a proxy", () => new HttpResponse(null, { status: 502 }), serverProblem],
    // openapi-fetch gives no error at all for a response with Content-Length 0.
    [
      "an empty body and Content-Length 0",
      () => new HttpResponse(null, { status: 502, headers: { "Content-Length": "0" } }),
      serverProblem,
    ],
    [
      "an HTML body from a proxy",
      () => HttpResponse.html("<h1>Bad gateway</h1>", { status: 502 }),
      serverProblem,
    ],
    ["a JSON null", () => HttpResponse.json(null, { status: 400 }), rejectedRequest],
    [
      "a detail that is not text",
      () => HttpResponse.json({ detail: { text: "x" } }, { status: 422 }),
      rejectedRequest,
    ],
    [
      "a detail that is empty",
      () => HttpResponse.json(problem(422, { detail: "" }), { status: 422 }),
      rejectedRequest,
    ],
  ])(
    "shows a general message and no success message for a failed response with %s",
    async (_name, respond, message) => {
      server.use(http.post(registerUrl, respond));
      renderApp("/register");

      await fillAndSubmit(registerUserRequest());

      expect(await screen.findByText(message)).toBeInTheDocument();
      expect(screen.getByRole("alert")).toHaveTextContent(message);
    },
  );

  it("shows a message when the server does not answer", async () => {
    server.use(http.post(registerUrl, () => HttpResponse.error()));
    renderApp("/register");

    await fillAndSubmit(registerUserRequest());

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "The application cannot reach the server.",
    );
  });

  it("removes the old message when the person sends the form again", async () => {
    server.use(
      http.post(
        registerUrl,
        () =>
          HttpResponse.json(problem(409, { detail: "The email address is already registered." }), {
            status: 409,
          }),
        { once: true },
      ),
      ...registrationHandlers([]),
    );
    renderApp("/register");
    const person = userEvent.setup();
    await fillAndSubmit(registerUserRequest());
    await screen.findByText("The email address is already registered.");

    await person.click(screen.getByRole("button", { name: "Register" }));

    expect(await screen.findByRole("heading", { level: 1, name: "User" })).toBeInTheDocument();
    expect(screen.queryByText("The email address is already registered.")).not.toBeInTheDocument();
  });

  it("does not send the form twice while it waits for the server, and the button keeps the focus", async () => {
    let requests = 0;
    const ann = user();
    server.use(
      http.post(registerUrl, async () => {
        requests += 1;
        await delay(100);
        return HttpResponse.json(ann, { status: 201 });
      }),
      http.get<{ id: string }, never, User>(userUrl, () => HttpResponse.json(ann)),
    );
    renderApp("/register");
    const person = userEvent.setup();
    const request = registerUserRequest();
    await person.type(screen.getByLabelText("Email address"), request.emailAddress);
    await person.type(screen.getByLabelText("Display name"), request.displayName);

    await person.click(screen.getByRole("button", { name: "Register" }));
    const button = screen.getByRole("button", { name: "Register" });
    expect(button).toHaveAttribute("aria-disabled", "true");
    expect(button).toHaveFocus();
    await person.keyboard("{Enter}");
    await person.click(button);

    await screen.findByRole("heading", { level: 1, name: "User" });
    expect(requests).toBe(1);
  });

  it("works with the keyboard only", async () => {
    const received: RegisterUserRequest[] = [];
    server.use(...registrationHandlers(received));
    renderApp("/register");
    const person = userEvent.setup();

    await person.tab();
    expect(screen.getByLabelText("Email address")).toHaveFocus();
    await person.keyboard("ann@example.com");
    await person.tab();
    expect(screen.getByLabelText("Display name")).toHaveFocus();
    await person.keyboard("Ann{Enter}");

    await screen.findByRole("heading", { level: 1, name: "User" });
    expect(received).toEqual([{ emailAddress: "ann@example.com", displayName: "Ann" }]);
  });

  it("tells the browser that the display name is a nickname, and not the full name of the person", () => {
    renderApp("/register");

    expect(screen.getByLabelText("Display name")).toHaveAttribute("autocomplete", "nickname");
    expect(screen.getByLabelText("Email address")).toHaveAttribute("autocomplete", "email");
  });
});
