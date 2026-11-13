# Project state

Scope: Read and write project facts and events with Cortex Project Memory.

- At session start, call `cortex_context({})` to inspect recent events and active `@context`, and confirm its notebook directory matches the intended project before saving.
- Treat worktree notebooks as separate; reconnect to the intended project's Cortex server rather than saving unrelated state in the current notebook.
- Find with `cortex_find`; read full records with `cortex_read`, following cursors.
- **Distinguish Facts vs. Events strictly:**
  - **Facts (`cortex_fact_put`)** are durable, present system invariants (*"what is currently true about this codebase"* — e.g. build commands, architecture topology, external service dependencies, `@context`). They use semantic concept IDs and can be updated via CAS or retired with tombstones. Never put transient actions, completed tasks, or historical log entries as facts.
  - **Events (`cortex_event_append`)** are chronological milestone outcomes (*"what just happened"* — e.g. MRs/PRs merged, bug investigations, test runs, CI results). Events are immutable, append-only, and automatically budget into the rolling context window.
- Save verified facts and evidence with `cortex_fact_put`; use the current `expected_revision` (`""` for new facts), rereading on conflicts.
- **MANDATORY MILESTONE LOGGING**: You MUST append an immutable milestone event via `cortex_event_append` immediately upon completing:
  - **Investigations & Diagnoses**: Findings, root causes identified, architectural trade-offs, or hypothesis disproofs (no git commit needed, but event record is required).
  - **Verified Code Slices**: Completed slices, bug fixes, and refactors (commit to git AND log event with commit SHA).
  - **Validation Milestones**: Test suite runs, benchmarks, MR/PR creations, and CI resolutions.
  Do not wait for user prompting or session end. An investigation without code changes is incomplete until recorded via `cortex_event_append`.
- Include evidence (e.g. commit hash, test outputs, URLs) and distinguish verified results from pending work.
- Retry writes with the same `operation_id` and identical input.
- Verify successful writes by reading the saved records back; an attempted write is not proof of persistence.
- Before finishing, save any remaining new facts and completed outcomes; never store secrets.
- If persistence is blocked, tell the user what remains unsaved before finishing; never claim a failed write was saved.
