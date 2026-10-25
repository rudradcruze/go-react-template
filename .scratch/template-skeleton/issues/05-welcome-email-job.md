# 05: Welcome email job

**Parent:** `.scratch/template-skeleton/spec.md`

**What to build:** After a Registration, the application sends a Welcome email to the new User as a background job. In this template, the mailer only writes the email to the log. A developer can see the jobs in Asynqmon.

**Blocked by:** 03 (Registration API)

**Status:** ready-for-agent

- [ ] After the Registration transaction commits, the API enqueues a `user:send_welcome_email` task on Redis DB 1, with a maximum of 3 retries.
- [ ] If the enqueue fails, the API logs the error, and the Registration still returns 201.
- [ ] The worker process runs the task. It sends the Welcome email through a mailer port. The mailer adapter writes the email to the log.
- [ ] `make dev` also starts the worker.
- [ ] `make up` also starts Asynqmon on port 5650.
- [ ] Unit tests cover the enqueue after Registration, the enqueue failure and the task handler, with mocks.
- [ ] An E2E test proves that one Registration puts one Welcome email task in the queue.
