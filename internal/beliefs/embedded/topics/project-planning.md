# Project planning and task breakdown

Scope: Converting a complex project, RFC, or engineering blueprint into actionable, codebase-grounded tickets or implementation tasks.

## Phase 1: Mandatory codebase grounding

Inspect Cortex Project Memory and the repository before proposing any tickets.

1. Query `cortex_find(kind: "fact", query: "<domain>")` and call `cortex_read` only on matching IDs to reuse already-verified architecture, seams, and consumer facts.
2. Inspect the repository with search and file-reading tools to identify the exact files, modules, APIs, schemas, and tests touched.
3. Note existing patterns and utilities to reuse, plus deprecated patterns or downstream consumers to protect.
4. Persist newly discovered **durable** codebase invariants (module boundaries, shared wrappers, consumer constraints) via `cortex_fact_put` with scoped `topics` and `evidence`.
5. Never dump raw file contents, throwaway scratch notes, or individual ticket lists into Facts; retire any temporary planning facts via `cortex_fact_retire` once tickets are finalized.
6. If any part of the blueprint conflicts with how the codebase actually works, STOP, record the conflict summary via `cortex_event_append`, and list those conflicts plus open questions before generating tickets.

## Phase 2: Sizing and anti-pitfall rules

- **Target granularity:** Aim for ~12–20 coherent tickets (hard cap: 25) for large projects, scaling proportionally for smaller scopes without manufacturing filler.
- **One reviewable PR per ticket:** Each ticket must map to one PR (~200–400 LOC, reviewable in under 1 hour).
- **Two sizing gates (every ticket MUST pass both):**
  1. *One-Sentence Diff Test:* You can describe the code change in one clear sentence.
  2. *Independent Verifiability:* The ticket produces a testable outcome (unit/integration test, passing migration, or working endpoint) without waiting on the next ticket.
- **No micro-tickets:** Bundle trivial related edits (e.g., DB column + model struct + regenerated types) into a single coherent ticket.
- **No discovery or spike tickets:** Answer codebase questions yourself in Phase 1. Only create a spike if it depends on an external system or vendor not visible in the repo, and explicitly justify why.

## Phase 3: Output schema (strict)

Output each ticket using ONLY these fields (keep each ticket concise—no narrative fluff or speculative risks):

### [TICKET-N] <Verb> <Concrete Target> so that <Outcome>
- **Goal:** One sentence stating what changes and why.
- **In Scope:** Bullet list of exact deliverables.
- **Out of Scope:** 1–2 explicit boundaries so the implementer doesn't overreach.
- **Codebase Pointers (Verified):** Exact `path/to/file` and symbol names read in Phase 1, plus the existing pattern to follow (mark any new files explicitly as `[NEW]`).
- **Acceptance Criteria:** 2–4 concrete, testable pass/fail checks (commands, test cases, or behavior).
- **Depends On:** `[TICKET-X]` or `None`.

After finalizing the breakdown, record a concise milestone via `cortex_event_append` (ticket count and scope summary, not full ticket bodies).
