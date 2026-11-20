# Project planning and task breakdown (Stage 3)

Scope: Slicing an accepted RFC (with empty Section 7) into actionable, PR-sized tickets or Jira tasks.

## Gate 2 prerequisite & fresh-context Cortex retrieval

- **Gate 2 hard stop:** Refuse to slice if RFC Section 7 (`Open Questions`) has unresolved items or `Status != Accepted`.
- **Fresh-context grounding (anti-rot):** Do not carry stale Stage 1 exploration logs into Stage 3. Ground from the accepted RFC + targeted `cortex_find(kind: "fact", query: "<domain>")` / `cortex_read` calls only.
- Inspect the repo only to verify exact `file#Symbol` pointers not already covered by the RFC or Cortex facts.
- Never dump raw file contents, scratch notes, or ticket lists into Facts; retire any temporary planning facts via `cortex_fact_retire` once finalized.

## Phase 2: Sizing and anti-pitfall rules

- **Target granularity:** Aim for ~12–20 coherent tickets (hard cap: 25) for large projects, scaling proportionally for smaller scopes without manufacturing filler.
- **One reviewable PR per ticket:** Each ticket must map to one PR (~200–400 LOC, reviewable in under 1 hour), aligned with the RFC's Section 6 vertical milestones.
- **Two sizing gates (every ticket MUST pass both):**
  1. *One-Sentence Diff Test:* You can describe the code change in one clear sentence.
  2. *Independent Verifiability:* The ticket produces a testable outcome (unit/integration test, passing migration, or working endpoint) without waiting on the next ticket.
- **No micro-tickets:** Bundle trivial related edits (e.g., SQL migration + domain struct + store methods) into a single coherent ticket.
- **No discovery or spike tickets:** Answer codebase/doc questions in Stage 1. Only create a spike if an external runtime test cannot be determined from the repo or docs, with explicit justification.

## Phase 3: Output schema & Gate 3 dry-run

- **Gate 3 (Dry-run before Jira writes):** Output proposed tickets in Markdown (or a local JSON manifest) for human review before invoking any remote Jira issue-creation APIs.
- **Traceability:** Embed direct links to the RFC and relevant `docs/adr/` files in ticket descriptions where applicable.

Output each ticket using ONLY these fields (no narrative fluff or speculative risks):

### [TICKET-N] <Verb> <Concrete Target> so that <Outcome>
- **Goal:** One sentence stating what changes and why.
- **In Scope:** Bullet list of exact deliverables.
- **Out of Scope:** 1–2 explicit boundaries derived from RFC Non-Goals.
- **Codebase Pointers (Verified):** Exact `path/to/file` and symbol names, plus existing pattern to follow (mark new files `[NEW]`).
- **Acceptance Criteria:** 2–4 concrete, testable pass/fail checks (commands, test cases, or behavior).
- **Depends On:** `[TICKET-X]` or `None`.

After finalizing the breakdown, record a concise milestone via `cortex_event_append` (ticket count and scope summary, not full ticket bodies).
