# RFC authoring and agent-ready blueprints

Scope: Authoring or reviewing an agent-ready RFC (Request for Comments) before `project-planning` ticket slicing.

## Mandatory pre-write grounding & Cortex memory

- **Inspect the repo BEFORE drafting Section 4:** Query `cortex_find(kind: "fact", query: "<domain>")` and `cortex_read` matching IDs first, then read actual migrations/schemas, routes/handlers, and workers in the codebase. Never guess or omit table, column, or route names.
- Persist newly discovered **durable** architectural invariants or accepted ADR decisions via `cortex_fact_put` with scoped `topics` and `evidence`; never dump raw file contents or draft RFC prose into Facts.
- Record RFC lifecycle milestones (e.g., Draft -> Accepted, Section 7 cleared) via `cortex_event_append`.

## Concreteness gate (ban vague prose)

Every RFC MUST specify exact identifiers verified from the codebase rather than abstract summaries:
- **Schema / Storage:** Name exact tables/collections, columns, types, defaults, enum values, and indexes (e.g., `webhook_deliveries`: `attempt_count` int default 0, `next_retry_at` timestamptz, partial index `(status, next_retry_at) WHERE status = 'retrying'`).
- **API / Events:** Name exact HTTP method + path (or event/topic name), status codes, and state transitions (e.g., `POST /admin/v1/webhooks/:id/replay` -> `202 Accepted` or `409 Conflict`).
- **Rollout / Config:** Name exact feature flags or config keys (e.g., `WEBHOOK_RETRIES_ENABLED`).
- **Reject generic placeholders:** Never write `"update the database"`, `"add an endpoint"`, or `"handle retries"` without exact table, column, route, and state-transition names.

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

### Non-Goals (Explicitly Out of Scope)
- [What we are NOT building—feeds directly into ticket `Out of Scope` guards]

## 3. Codebase Anchors (Verified)
- **Primary entry points / modules:** `internal/webhooks/dispatcher.go`, `internal/webhooks/store.go`
- **Existing patterns to reuse:** [e.g., "Follow `pkg/retry/exponential.go` and `internal/api/audit_handler.go`"]
- **Deprecated patterns to avoid:** [e.g., "Do not use `pkg/legacyqueue/poller.go`"]
- **Downstream consumers / contracts affected:** [e.g., "`billing-metrics` reads `webhook_deliveries.status`; keep `delivered` and `failed`"]

## 4. Proposed Design
- **Schema / Storage:** [Exact table, columns, types, defaults, enum values, and indexes—e.g., add `attempt_count` (int, default 0), `next_retry_at` (timestamptz, nullable), `last_error` (text) to `webhook_deliveries`; partial index on `(status, next_retry_at)` where `status = 'retrying'`]
- **Worker / State Flow:** [Exact state transitions and helper calls—e.g., on non-2xx/timeout, if `attempt_count < 5`, set `status = 'retrying'`; on attempt 5 failure, set `status = 'dead_letter'`]
- **API / Events:** [Exact method, path, payload, and status codes—e.g., `POST /admin/v1/webhooks/:id/replay` resets `dead_letter` -> `retrying` with `attempt_count = 0` (`202` / `409`)]

## 5. Alternatives Considered & Rejected
- **[Option B]:** Rejected because [concrete trade-off, e.g., latency, operational cost, or coupling].

## 6. Rollout & Verification Strategy
- **Feature flags / rollout:** [Exact flag name + rollout order, e.g., ship migration first, enable `WEBHOOK_RETRIES_ENABLED` in staging]
- **Vertical milestones (3–5 slicing boundaries):**
  1. [Milestone 1: e.g., Schema migration + store methods pass DB integration tests]
  2. [Milestone 2: e.g., Dispatcher retry/DLQ state machine verified with mock HTTP server tests]
  3. [Milestone 3: e.g., Admin replay endpoint + metrics verified end-to-end]

## 7. Open Questions (Must Be Empty Before Jira Slicing)
- [x] [Question 1] -> **Decision:** [Locked answer]
- [ ] [Question 2] -> **Owner:** @[name] (Resolve before running `project-planning`)
```
