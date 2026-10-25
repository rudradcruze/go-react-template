import { randomUUID } from "node:crypto";

import type { Page } from "@playwright/test";

export interface Person {
  emailAddress: string;
  displayName: string;
}

// Each test makes its own data with unique values. A test does not use the data of a different test,
// so the tests can run in any order, also at the same time.
export function newPerson(): Person {
  const id = randomUUID().slice(0, 8);
  return { emailAddress: `e2e-${id}@example.com`, displayName: `Person ${id}` };
}

// fillRegistrationForm enters the values of a person in the Registration page, and sends the form.
export async function fillRegistrationForm(page: Page, person: Person) {
  await page.getByLabel("Email address").fill(person.emailAddress);
  await page.getByLabel("Display name").fill(person.displayName);
  await page.getByRole("button", { name: "Register" }).click();
}
