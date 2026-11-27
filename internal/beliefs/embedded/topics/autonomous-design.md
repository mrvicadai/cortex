# Autonomous end-to-end design (ADR -> Feasibility -> RFC -> Tasks)

Scope: Default end-to-end design pipeline (`ADR -> feasibility-audit -> rfc-authoring -> project-planning`) using Via Negativa and best-judgment gate clearance.

## Pipeline & HITL overlay

- Call `cortex_get_belief(topics: ["feasibility-audit", "rfc-authoring", "project-planning"])` and enforce all quality, schema, concreteness, and sizing rules across Stages 1–3.
- Pause for human signoff only when the user explicitly requests checkpoints or `hitl-design-gates` is active.

## Stage 0: ADR and Via Negativa framing

- Run the **Via Negativa Checklist** before proposing any new component:
  1. **Subtract first:** Solve the problem by deleting broken checks, dead code, or redundant steps before adding new ones.
  2. **Reuse primitives:** Reuse existing tables, endpoints, workers, helpers, or Cortex facts before creating new ones.
  3. **Minimum reversible surface:** Avoid new infrastructure, dependencies, or feature flags unless simpler options are provably insufficient.
- Compare at most 2–3 viable options, select the simplest reversible one using best judgment, and record the choice in `docs/adr/NNNN-title.md` (for hard-to-reverse choices) or a Stage 0 ADR summary.

## Best-judgment tie-breaker & critic pass

- Resolve open design questions using this ordered rubric: **1. Via Negativa** (least new code/infra) -> **2. Reversibility** (easiest to change later) -> **3. Repo precedent** (matches existing patterns).
- Before clearing Stage 2, verify every cited `file#Symbol` exists in the repo and run an adversarial critic check (what can be deleted or reused, and what fails under concurrency or partial failure).
- Halt and escalate *only* on true external dead-ends: missing credentials/permissions, private undocumented APIs, or contradictory goals.

## Anti-rot Cortex checkpoints

- Persist durable architectural decisions and verified constraints via `cortex_fact_put` at Stage 0 and Stage 1.
- Append a concise milestone via `cortex_event_append` at each stage transition (`Stage 0 ADR locked`, `Stage 1 Feasibility cleared`, `Stage 2 RFC accepted`, `Stage 3 Tasks sliced`).
- Ground each stage from the prior stage's structured artifact and targeted `cortex_find` / `cortex_read` calls rather than raw exploration logs.
