import { expect, test } from "@playwright/test";

import { fillRegistrationForm, newPerson } from "./support";

test("a person registers and sees the User page with the email address and the display name", async ({
  page,
}) => {
  const person = newPerson();
  await page.goto("/");
  await page.getByRole("link", { name: "Register" }).click();
  await fillRegistrationForm(page, person);

  await expect(page).toHaveURL(/\/users\/[0-9a-f-]{36}$/);
  await expect(page.getByText(person.emailAddress)).toBeVisible();
  await expect(page.getByText(person.displayName)).toBeVisible();
});

test("after a Registration, the focus is on the heading of the User page and the tab has its title", async ({
  page,
}) => {
  await page.goto("/register");
  await fillRegistrationForm(page, newPerson());

  await expect(page.getByRole("heading", { level: 1, name: "User" })).toBeFocused();
  await expect(page).toHaveTitle("User - go-react-template");
});

test("the User page shows a User that a person opens by its URL", async ({ page, request }) => {
  const person = newPerson();
  const created = await request.post("/api/v1/users", { data: person });
  expect(created.status()).toBe(201);
  const { id } = await created.json();

  // nginx gives the page of the frontend for this path, and the page reads the User from the API.
  await page.goto(`/users/${id}`);

  await expect(page.getByText(person.emailAddress)).toBeVisible();
  await expect(page.getByText(person.displayName)).toBeVisible();
});

test("a person who enters an email address that is registered, with other letter case, sees the error", async ({
  page,
  request,
}) => {
  const person = newPerson();
  const created = await request.post("/api/v1/users", { data: person });
  expect(created.status()).toBe(201);

  await page.goto("/register");
  await fillRegistrationForm(page, {
    emailAddress: person.emailAddress.toUpperCase(),
    displayName: "Another name",
  });

  await expect(page.getByRole("alert")).toHaveText(
    "The email address is already registered. Use a different email address.",
  );
  await expect(page).toHaveURL(/\/register$/);
});

test("a person who enters an email address that is not valid sees the error", async ({ page }) => {
  await page.goto("/register");
  await fillRegistrationForm(page, { emailAddress: "not-an-address", displayName: "Someone" });

  await expect(page.getByRole("alert")).toHaveText(
    "The email address is not valid. Enter an address in the format name@example.com.",
  );
});

test("a person who opens the URL of a User that does not exist sees the message of the API", async ({
  page,
}) => {
  await page.goto("/users/0198f2a0-0000-7000-8000-000000000000");

  await expect(page.getByRole("alert")).toHaveText("The User does not exist. Check the ID.");
});
