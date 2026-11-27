# Human-in-the-loop design gates (HITL overlay)

Scope: Additive human signoff checkpoints on top of `autonomous-design`, `feasibility-audit`, `rfc-authoring`, and `project-planning` when the user explicitly requests review gates or `/plan` signoff.

## Mandatory human pause points

- **Gate 1 (Post-Stage 1):** After producing the `feasibility-audit` Dependency & Capability Matrix, halt and wait for human approval before starting `rfc-authoring`, even if all dependencies are `Verified`.
- **Gate 2 (Post-Stage 2):** Leave ambiguous Section 7 `Open Questions` unchecked (`[ ]`) with a named owner and halt until the human marks `Status: Accepted` before starting `project-planning`.
- **Gate 3 (Post-Stage 3):** Output the proposed task breakdown in Markdown or JSON for human review before queueing tasks or invoking remote Jira/Linear APIs.
