Scope: Cross-repository MR/PR guidance, including non-interaction, non-destructive conflicts, description shape, and notification restrictions.

## Irrefutable proof
Produce irrefutable, directly observed proof of runtime behavior before opening or updating an MR/PR under [testing guidance](testing.md).
Include exact commands, observed outcomes, and relevant outputs or links in the MR/PR description.
Disclose deferred or environment-gated validation explicitly rather than implying unperformed coverage.

## Default PR/MR babysitting
After opening or pushing to an MR/PR, record the MR/PR URL via `cortex_event_append` (or `cortex_todo_done` with the URL in `evidence`), and default to babysitting it until CI is green and actionable feedback is resolved, unless told to stop.
Wait for pending CI checks, inspect failures, conflicts, and review or bot comments, push verified fixes, and re-verify.
Stop only when CI is green with only human approval or explicit user action remaining, or when blocked on permissions or user input.

## Authorized user actions
Do not comment on, approve, merge, close, or force-push an MR unless the user explicitly confirms that named action.

## Findings and thread closure
Address findings in code, push, and explain in chat; humans close threads.
Treat bot findings as empirically verifiable evidence, including smells not fixable as described.

## Other-party surfaces and denied writes
A general go-ahead cannot authorize a surface owned by another party.
Ask once for the named edit and never retry a denied write through another transport.

## Ticket-prefixed title
Use a ticket-prefixed MR title, never a conventional-commit prefix.

## Load-bearing description
Start with a brief bold **TL;DR** of what changes and why, followed by short walkthrough cards separated by `---` (one idea per card with a descriptive heading or bold takeaway).
Explain the flow, choices, checks, and material limits in plain language with small concrete examples, avoiding dry summaries and unexplained jargon.
Use small, validated Mermaid diagrams with plain labels for complex flows, dependencies, or concurrency.
Use descriptive names instead of opaque IDs like `I01` or `I02`, preserve inline bot summaries, and allow an optional trailing ticket link.

## Preserve MR metadata
Check `cortex_find(kind: "fact", query: "mr")` for recorded repository MR conventions before setting labels or reviewers.
Use low-impact labels unless impact is genuinely higher.
Assign the documented owner, preserve existing reviewers, and honor any explicit replacement list.
Only add reviewers, assignees, or @mentions when explicitly requested; creation, edits, and green CI do not authorize notification tags.

## Verify command outcomes
Specify the source branch, use supported noninteractive commands, read metadata back, project API responses, and use shell API calls where the MCP cannot return collection endpoints.

## Human approval gate
Treat a green pipeline with only human approval pending as complete from the agent side.
Trust build and vet over unreliable worktree language-server diagnostics.

## Fast pipeline-covered iteration
For small reviewed pipeline-covered changes, commit, push, open the MR, inspect CI, then add follow-up commits without blocking on redundant slow local checks.
Run near-zero-cost formatting and build-drift checks first; verify locally before push only when CI cannot answer the risk.

## Protect generated artifacts
Never commit while code generation or build work is in flight.
Inspect git status immediately before staging because interrupted generation can delete artifacts.
