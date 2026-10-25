# Writing rule: ASD-STE100

Write all human-readable text in this repo in ASD-STE100 Simplified Technical English (STE).

## Scope

STE applies to all text that a person reads:

- UI text: labels, buttons, placeholders, empty states, toasts
- Success and error messages in the frontend and in API responses
- Validation messages, log messages and CLI output
- Docs, `README.md`, `CONTEXT.md`, ADRs and issues in `.scratch/`
- Code comments, commit messages and PR text

Code identifiers, config keys and text copied from third parties (for example `.claude/skills/`) keep their original form.

## Rules

- Use one word for one meaning. Use the STE dictionary meaning of each word. Use the same term for the same thing every time, and take domain terms from `CONTEXT.md`.
- Write sentences of 20 words or fewer in instructions, and 25 words or fewer in descriptions.
- Write paragraphs of 6 sentences or fewer.
- Write instructions in the imperative: "Enter your email address."
- Write one instruction in each sentence, unless the user does the two actions at the same time.
- Use the active voice. Name the person or thing that does the action.
- Use only the simple tenses: present, past and future.
- Use the "-ing" form only in technical names.
- Use single-word verbs ("start", "stop", "remove") in place of phrasal verbs ("fire up", "shut down", "take out").
- Write full forms: "do not", "cannot", "it is".
- Keep the articles ("a", "the") and all the words that make a sentence complete.
- Use noun clusters of 3 words or fewer.
- Use a vertical list for steps and for more than 3 items.
- Write a warning as a command first, then give the reason: "Do not delete this key. The app cannot decrypt old data without it."

## Success and error messages

- Success: state the result. "The profile is saved."
- Error: state the problem, then tell the user what to do. "The email address is not valid. Enter an address in the format name@example.com."
- Give the specific cause. A general message does not help the user to correct the problem.

| Write | In place of |
| --- | --- |
| The file is saved. | File saved successfully! |
| The password is incorrect. Enter the password again. | Oops! Looks like your password's wrong. |
| The server did not respond. Try again in 1 minute. | Something went wrong. |
| Connect the database before you start the server. | Make sure you've got the DB hooked up before firing up the server. |

## Check

Before you complete a change, read each new or changed text string. Correct each string that does not obey these rules.
