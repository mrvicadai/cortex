# Autonomous end-to-end design (ADR -> Feasibility -> RFC -> Tasks)

Scope: Orchestrating an autonomous end-to-end design from Stage 0 ADR through Stage 3 task planning when explicitly instructed, overriding the default HITL gates in `feasibility-audit`, `rfc-authoring`, and `project-planning`.

## Precedence and mode activation

- Apply this workflow only when the user asks for an autonomous or end-to-end design; otherwise keep default HITL gates.
- Call `cortex_get_belief(topics: ["feasibility-audit", "rfc-authoring", "project-planning"])` and enforce all their quality, schema, concreteness, and sizing rules except their human-approval hard stops.

## Stage 0: ADR and Via Negativa framing

- Run the **Via Negativa Checklist** before proposing any new component:
  1. **Subtract first:** Solve the problem by deleting broken checks, dead code, or redundant steps before adding new ones.
  2. **Reuse primitives:** Reuse existing tables, endpoints, workers, helpers, or Cortex facts before creating new ones.
  3. **Minimum reversible surface:** Avoid new infrastructure, dependencies, or feature flags unless simpler options are provably insufficient.
- Compare at most 2–3 viable options, select the simplest reversible one using best judgment, and record the choice in `docs/adr/NNNN-title.md` (for hard-to-reverse choices) or a Stage 0 ADR summary.

## Autonomous gate overrides (Stages 1–3)

- **Stage 1 (`feasibility-audit`):** Self-clear Gate 1 when critical dependencies are `Verified` or have a verified *via negativa* workaround.
- If a critical dependency is `Missing` or `Unverified`, loop back to Stage 0 to choose a verified simpler path; halt and escalate *only* on true external dead-ends (missing credentials/permissions, inaccessible private APIs, or contradictory goals).
- **Stage 2 (`rfc-authoring`):** Link the Stage 0 ADR, record rejected additive options in Section 5 (`Alternatives Considered & Rejected`), resolve every Section 7 `Open Question` using best judgment and codebase evidence, and set `Status: Accepted (Autonomous)`.
- **Stage 3 (`project-planning`):** Slice the accepted RFC into PR-sized tasks, self-clear Gate 3 for Markdown/JSON output (and queue slices via `cortex_todo_add` if execution is requested), and keep remote Jira/Linear writes gated unless explicitly authorized.

## Anti-rot Cortex checkpoints

- Persist durable architectural decisions and verified constraints via `cortex_fact_put` at Stage 0 and Stage 1.
- Append a concise milestone via `cortex_event_append` at each stage transition (`Stage 0 ADR locked`, `Stage 1 Feasibility cleared`, `Stage 2 RFC accepted`, `Stage 3 Tasks sliced`).
- Ground each stage from the prior stage's structured artifact and targeted `cortex_find` / `cortex_read` calls rather than raw exploration logs.
