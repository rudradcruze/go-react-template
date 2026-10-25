# 07: Registration page

**Parent:** `.scratch/template-skeleton/spec.md`

**What to build:** A person registers on the frontend with an email address and a display name. The page tells the person when the Registration is successful. When the API rejects the Registration, the page shows the reason from the API.

**Blocked by:** 03 (Registration API), 06 (Frontend skeleton)

**Status:** ready-for-agent

- [ ] The Registration page has a form with an email address field and a display name field. Each field has a label.
- [ ] The form sends the Registration to the API through the typed client.
- [ ] After a successful Registration, the page shows a success message.
- [ ] After a 409 or 422 response, the page shows the `detail` text of the problem response.
- [ ] A person can use the form with only the keyboard.
- [ ] All page text obeys the STE rule.
- [ ] Unit tests use MSW and fixture builders. They cover the success, the 409 response and the 422 response.
