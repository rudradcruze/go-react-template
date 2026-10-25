## Agent skills

### Issue tracker

The issues are local markdown files in `.scratch/<feature>/`. See `docs/agents/issue-tracker.md`.

### Triage labels

The triage labels are the 5 default labels: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human` and `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

This repo has one context: one `CONTEXT.md` file and one `docs/adr/` folder at the repo root. See `docs/agents/domain.md`.

### Code Modification Principles

When modifying code, do not add unnecessary comments, abstractions, wrappers, headings, documentation, or explanatory prose merely to make the change look polished. Follow the repository's existing naming, formatting, and architectural conventions. Make the smallest clear change that solves the actual problem. Write all prose to the ASD-STE100 rules in `.claude/rules/writing.md` (no em dashes).