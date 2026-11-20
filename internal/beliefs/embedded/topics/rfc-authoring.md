# RFC authoring and agent-ready blueprints

Scope: Authoring or reviewing an agent-ready RFC (Request for Comments) before `project-planning` ticket slicing.

## Cortex memory integration (anti-bloat)

- Before populating Section 3, query `cortex_find(kind: "fact", query: "<domain>")` and call `cortex_read` only on matching IDs to reuse verified module boundaries, patterns, and consumers.
- Persist newly discovered **durable** architectural invariants or accepted ADR-level decisions via `cortex_fact_put` with scoped `topics` and `evidence`; never dump raw file contents or draft RFC prose into Facts.
- Record RFC lifecycle milestones (e.g., Draft -> Accepted, Section 7 cleared) via `cortex_event_append`.

## Three rules for agent-ready RFCs

1. **Keep Section 7 empty before slicing:** Treat unresolved items in Section 7 as a hard blocker before running `project-planning`, eliminating discovery/spike tickets.
2. **Include both "reuse" and "avoid" in Section 3:** Naming exact helpers to copy and deprecated patterns to avoid cuts file-reading overhead and prevents reinventing utilities.
3. **Use Section 6's vertical milestones as slicing boundaries:** Give `project-planning` 3–5 high-level checkpoints so it produces ~12–20 coherent, PR-sized tickets instead of micro-tasks.

## RFC template (strict)

```markdown
# RFC-[NNN]: [Short Feature or System Title]
- **Status:** Draft | In Review | Accepted
- **Owner:** @[name]
- **Related ADRs:** [None | `docs/adr/000X-...md`]

## 1. Context & Problem
[2–4 sentences explaining what is broken, missing, or changing, and why now.]

## 2. Goals & Non-Goals
### Goals (In Scope)
- [Measurable outcome 1]
- [Measurable outcome 2]

### Non-Goals (Explicitly Out of Scope)
- [What we are NOT building in this project—feeds directly into ticket `Out of Scope` guards]
- [Adjacent refactors or future phases to leave untouched]

## 3. Codebase Anchors (Verified)
[Populate in read-only mode before finalizing the RFC:]
- **Primary entry points / modules:** `path/to/service.ts`, `path/to/router.go`
- **Existing patterns to reuse:** [e.g., "Follow retry middleware in `pkg/http/retry.go`"]
- **Deprecated patterns to avoid:** [e.g., "Do not use legacy `v1/client.ts`"]
- **Downstream consumers / contracts affected:** [e.g., "`billing-worker` consumes `OrderCreated` events"]

## 4. Proposed Design
[Describe architecture, data flow, schema/API changes, and failure modes. Focus on contracts and boundaries, not line-by-line code.]

### Key Interfaces / Data Changes
- **Schema / Storage:** [New tables, columns, or migrations]
- **API / Events:** [Endpoints, request/response shapes, or queue payloads]

## 5. Alternatives Considered & Rejected
- **[Option B]:** Rejected because [concrete trade-off, e.g., latency, operational cost, or coupling].
*(Promote any major, hard-to-reverse choice here to a standalone ADR in `docs/adr/`.)*

## 6. Rollout & Verification Strategy
- **Feature flags / rollout:** [e.g., dark launch behind `enable_new_sync`, backfill script, then 10% -> 100%]
- **Vertical milestones (how we verify progress):**
  1. [Milestone 1: e.g., Schema + repository layer passes integration tests]
  2. [Milestone 2: e.g., Worker processes events idempotently in staging]
  3. [Milestone 3: e.g., API endpoint + UI wired behind feature flag]

## 7. Open Questions (Must Be Empty Before Jira Slicing)
- [x] [Question 1] -> **Decision:** [Locked answer]
- [ ] [Question 2] -> **Owner:** @[name] (Resolve before running `project-planning`)
```
