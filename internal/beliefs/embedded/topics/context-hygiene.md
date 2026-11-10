# Context hygiene

Scope: Reducing the token cost of shell/MCP calls and responses.

These instructions are aspirational unless an enforcing mechanism is named; offline Markdown validation checks links and headings, not compliance.

## Input and output bounds

Save tokens in what you read, what you emit, and what you let tools return; re-check point-in-time harness budgets before asserting them.
Use `git log --oneline -3`; inspect `git diff` / `git show --stat` first; use `2>&1 | tail -20` for build/test output; and bound recursive `find` / `ls` through `head`.
Section-read files, never whole-file, except where a canonical policy requires a whole-document read.

## MCP and REST

Aggregate MCP data and telemetry server-side; never fetch raw rows to count them yourself.
Name the 1–3 fields needed by an MCP query; never `[*]`.
Use the tightest search limit that answers the question and widen only if empty.
A bare REST endpoint is `[*]`: GitLab/Jira expose no field-selection parameter, so project with `jq` or `python3 -c` in the same call.
Use ToolSearch's fully-qualified name.
Load both required skills together when a server gates tools behind two skills.

## Side channels

Before adding a side channel, ask who reads it back and when; prefer channels the harness maintains passively.
