# Project state

Scope: Read and write project facts and events with Fieldnotes.

- Start with `fieldnotes_context({})`; confirm its notebook directory matches the intended project before saving.
- Treat worktree notebooks as separate; reconnect to the intended project's Fieldnotes server rather than saving unrelated state in the current notebook.
- Find with `fieldnotes_find`; read full records with `fieldnotes_read`, following cursors.
- Save verified facts and evidence with `fieldnotes_fact_put`; use the current `expected_revision` (`""` for new facts), rereading on conflicts.
- Proactively append meaningful completed outcomes with `fieldnotes_event_append`, without waiting for a user request.
- Record milestones as they complete, not only at session end: investigations, changes, validation results, MR/CI milestones, and review resolutions.
- Include evidence and distinguish verified results from pending work.
- Retry writes with the same `operation_id` and identical input.
- Verify successful writes by reading the saved records back; an attempted write is not proof of persistence.
- Before finishing, save any remaining new facts and completed outcomes; never store secrets.
- If persistence is blocked, tell the user what remains unsaved before finishing; never claim a failed write was saved.
