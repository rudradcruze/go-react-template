# 08: User page

**Parent:** `.scratch/template-skeleton/spec.md`

**What to build:** A person sees the data of a User on the User page. After a successful Registration, the frontend opens the User page of the new User.

**Blocked by:** 04 (Get a User through the cache), 07 (Registration page)

**Status:** ready-for-agent

- [ ] The User page shows the email address and the display name of the User.
- [ ] After a successful Registration, the frontend opens the User page of the new User, in place of the success message from ticket 07.
- [ ] While the data loads, the page shows a loading state.
- [ ] When the API returns 404, the page shows a clear message that obeys the STE rule.
- [ ] Unit tests use MSW. They cover the User data, the loading state and the 404 response.
