import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { delay, http, HttpResponse } from "msw";
import { describe, expect, it } from "vite-plus/test";

import { problem } from "../../fixtures/problem";
import {
  registerUserRequest,
  user,
  type RegisterUserRequest,
  type User,
} from "../../fixtures/user";
import { server } from "../../mocks/server";
import { renderApp } from "../../render";

const registerUrl = "/api/v1/users";
const serverProblem = "The server cannot complete the request. Try again in 1 minute.";
const rejectedRequest = "The server did not accept the request. Check the values. Then try again.";

async function fillAndSubmit(request: RegisterUserRequest) {
  const person = userEvent.setup();
  await person.type(screen.getByLabelText("Email address"), request.emailAddress);
  await person.type(screen.getByLabelText("Display name"), request.displayName);
  await person.click(screen.getByRole("button", { name: "Register" }));
}

// registerHandler answers with a User, and it saves the body of the request.
function registerHandler(received: RegisterUserRequest[]) {
  return http.post<never, RegisterUserRequest, User>(registerUrl, async ({ request }) => {
    const body = await request.json();
    received.push(body);
    return HttpResponse.json(user(body), { status: 201 });
  });
}

describe("the Registration page", () => {
  it("has a label for each field and a button", () => {
    renderApp("/register");

    expect(screen.getByRole("heading", { level: 1, name: "Registration" })).toBeInTheDocument();
    expect(screen.getByLabelText("Email address")).toBeInTheDocument();
    expect(screen.getByLabelText("Display name")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Register" })).toBeInTheDocument();
  });

  it("has the regions of the messages before it has a message, so that a screen reader announces the first message", () => {
    renderApp("/register");

    expect(screen.getByRole("status")).toBeEmptyDOMElement();
    expect(screen.getByRole("alert")).toBeEmptyDOMElement();
  });

  it("sends the email address and the display name that the person entered", async () => {
    const received: RegisterUserRequest[] = [];
    server.use(registerHandler(received));
    renderApp("/register");
    const request = registerUserRequest({ emailAddress: "ann@example.com", displayName: "Ann" });

    await fillAndSubmit(request);

    await screen.findByText("The Registration is complete.");
    expect(received).toEqual([request]);
  });

  it("shows a success message and no error message after a Registration", async () => {
    server.use(registerHandler([]));
    renderApp("/register");

    await fillAndSubmit(registerUserRequest());

    expect(await screen.findByRole("status")).toHaveTextContent("The Registration is complete.");
    expect(screen.getByRole("alert")).toBeEmptyDOMElement();
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
    expect(screen.getByRole("status")).toBeEmptyDOMElement();
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
      expect(screen.getByRole("status")).toBeEmptyDOMElement();
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
      registerHandler([]),
    );
    renderApp("/register");
    const person = userEvent.setup();
    await fillAndSubmit(registerUserRequest());
    await screen.findByText("The email address is already registered.");

    await person.click(screen.getByRole("button", { name: "Register" }));

    expect(await screen.findByText("The Registration is complete.")).toBeInTheDocument();
    expect(screen.getByRole("alert")).toBeEmptyDOMElement();
  });

  it("does not send the form twice while it waits for the server, and the button keeps the focus", async () => {
    let requests = 0;
    server.use(
      http.post(registerUrl, async () => {
        requests += 1;
        await delay(100);
        return HttpResponse.json(user(), { status: 201 });
      }),
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

    await screen.findByText("The Registration is complete.");
    expect(requests).toBe(1);
  });

  it("works with the keyboard only", async () => {
    const received: RegisterUserRequest[] = [];
    server.use(registerHandler(received));
    renderApp("/register");
    const person = userEvent.setup();

    await person.tab();
    expect(screen.getByLabelText("Email address")).toHaveFocus();
    await person.keyboard("ann@example.com");
    await person.tab();
    expect(screen.getByLabelText("Display name")).toHaveFocus();
    await person.keyboard("Ann{Enter}");

    await screen.findByText("The Registration is complete.");
    expect(received).toEqual([{ emailAddress: "ann@example.com", displayName: "Ann" }]);
  });

  it("tells the browser that the display name is a nickname, and not the full name of the person", () => {
    renderApp("/register");

    expect(screen.getByLabelText("Display name")).toHaveAttribute("autocomplete", "nickname");
    expect(screen.getByLabelText("Email address")).toHaveAttribute("autocomplete", "email");
  });
});
