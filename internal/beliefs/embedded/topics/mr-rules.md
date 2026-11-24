Scope: Merge-request guidance, with repository boundaries defined below.

## Scope boundaries
Merge-request and pull-request guidance applies across repositories, including non-interaction, non-destructive conflicts, and the required description shape.
The notification-authorization restriction below also applies across repositories.

## Irrefutable proof
Produce irrefutable, directly observed proof that the changes work before opening or updating an MR/PR, following [testing guidance](testing.md).
Verify runtime behavior with concrete reproduction, before/after, or end-to-end checks rather than relying on build, vet, or unverified assumptions.
Include exact commands run, observed outcomes, and relevant links or outputs in the MR/PR description.
Disclose any deferred or environment-gated validation explicitly in the MR/PR description rather than implying unperformed coverage.

## Default PR/MR babysitting
After opening or pushing to an MR/PR, default to babysitting it until CI is green and actionable feedback is resolved, unless the user explicitly asks to stop.
Wait for CI/pipeline checks to reach a terminal state instead of declaring completion while checks are pending or queued.
Inspect failing checks, merge conflicts, and review or bot comments, fix verified issues in code, push follow-up commits, and re-verify.
Stop babysitting only when CI is green, all actionable findings are addressed, and only human approval or explicit user action remains, or when blocked on permissions or user input.

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
Start with a brief bold **TL;DR** describing what changes and why.
Follow it with short, plain-language walkthrough cards separated by horizontal rules (`---`), one idea per card, each with a descriptive heading or bold takeaway and a few short sentences.
Explain the flow, choices, checks, and material limits.
Use simple language and small, concrete examples to explain behavior; avoid dry implementation summaries and unexplained jargon.
Use small Mermaid diagrams to explain complex flows, dependencies, or concurrency. Keep labels plain and validate that diagrams render before publishing.
Use descriptive names instead of opaque inventory IDs such as `I01` or `I02`.
Preserve an inline bot summary when rewriting and allow an optional trailing ticket link.

## Preserve MR metadata
Use low impact labels unless impact is genuinely higher.
Assign the documented owner; preserve existing reviewers during unrelated MR edits.
Only add reviewers or @mention people when the user explicitly asks; creation, description edits and green CI do not authorize notification-triggering tags.

## Reviewer assignment
Only add reviewers or assignees when explicitly requested by the user.
Preserve existing reviewers and honor any explicit replacement list.

## Verify command outcomes
Specify the source branch, use supported noninteractive commands, read metadata back, project API responses, and use shell API calls where the MCP cannot return collection endpoints.

## Human approval gate
Treat a green pipeline with all actionable findings resolved and only human approval pending as complete from the agent side.
Trust build and vet over unreliable worktree language-server diagnostics.

## Fast pipeline-covered iteration
For small reviewed pipeline-covered changes, commit, push, open the MR, inspect CI, then add follow-up commits.
Do not block the writable branch on redundant slow local checks.

## Cheap checks first
Run near-zero-cost formatting and import/build-drift checks first.
Verify locally before push only when CI cannot answer the risk.

## Protect generated artifacts
Never commit while code generation or build work is in flight.
Inspect git status immediately before staging because interrupted generation can delete artifacts.
