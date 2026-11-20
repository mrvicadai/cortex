# RFC authoring and agent-ready blueprints

Scope: Authoring or reviewing an agent-ready RFC before `project-planning` ticket slicing.

## Mandatory pre-write grounding & Cortex memory

- **Inspect the repo BEFORE drafting Section 4:** Query `cortex_find(kind: "fact", query: "<domain>")` and `cortex_read` matching IDs first, then read actual migrations, schemas, routes, and workers. Never guess or omit table, column, or route names.
- Persist newly discovered **durable** invariants or accepted ADR decisions via `cortex_fact_put` (`topics` + `evidence`); never dump raw files or draft RFC prose into Facts.
- Record RFC lifecycle milestones via `cortex_event_append`.

## Concreteness & readability rules

- **No dense paragraphs:** Use short 1–2 sentence bullets, numbered steps, and Markdown tables instead of walls of text.
- **Simple language:** Write in plain, direct words; cut filler prose.
- **Explicit Schema & API:** Put table columns, types, defaults, indexes, HTTP methods/paths, payloads, and status codes in Markdown tables or typed/SQL/JSON blocks—never bury them in prose.
- **Ban vague placeholders:** Never write `"update the database"`, `"add an endpoint"`, or `"handle retries"` without exact identifiers (`webhook_deliveries`, `POST /admin/v1/webhooks/:id/replay`, `WEBHOOK_RETRIES_ENABLED`).
- **Three agent-readiness rules:** (1) Keep Section 7 empty before `project-planning` slicing; (2) Name both **reuse** and **avoid** patterns in Section 3; (3) Use Section 6's 3–5 vertical milestones as slicing boundaries.

## RFC template (strict)

```markdown
# RFC-[NNN]: [Short Feature or System Title]
- **Status:** Draft | In Review | Accepted  |  **Owner:** @[name]  |  **Related ADRs:** [None | `docs/adr/000X-...md`]

## 1. Context & Problem
[2–3 plain sentences explaining what is broken or missing, and why now.]

## 2. Goals & Non-Goals
- **Goals (In Scope):** [Measurable outcome 1]; [Measurable outcome 2]
- **Non-Goals (Out of Scope):** [What we are NOT building—feeds ticket `Out of Scope` guards]

## 3. Codebase Anchors (Verified)
- **Entry points / modules:** `internal/webhooks/dispatcher.go`, `internal/webhooks/store.go`
- **Reuse:** [Exact helper, e.g., `pkg/retry/exponential.go`, `internal/api/audit_handler.go`]
- **Avoid:** [Deprecated pattern, e.g., `pkg/legacyqueue/poller.go`]
- **Downstream consumers:** [e.g., "`billing-metrics` reads `webhook_deliveries.status`; keep `delivered` and `failed`"]

## 4. Proposed Design

### Schema / Storage
| Table | Column / Index | Type & Default | Notes |
| :--- | :--- | :--- | :--- |
| `webhook_deliveries` | `attempt_count`, `next_retry_at`, `last_error` | `int` (0), `timestamptz` (null), `text` | Add `retrying`, `dead_letter` to `status` |
| `webhook_deliveries` | `idx_webhooks_retry` (partial index) | `(status, next_retry_at)` | `WHERE status = 'retrying'` |

### Worker / State Flow
1. On non-2xx or timeout, if `attempt_count < 5`, set `status = 'retrying'`, increment `attempt_count`, and compute `next_retry_at` via `pkg/retry/exponential.go`.
2. On attempt 5 failure (or non-retryable 4xx), set `status = 'dead_letter'`.

### API / Events
| Endpoint / Event | Request / Behavior | Status Codes / Response |
| :--- | :--- | :--- |
| `POST /admin/v1/webhooks/:id/replay` | Reset `dead_letter` -> `retrying`, `attempt_count = 0` | `202 Accepted` (or `409 Conflict` if not `dead_letter`) |

## 5. Alternatives Considered & Rejected
| Option | Why Rejected |
| :--- | :--- |
| Kafka DLQ topic | Avoid operating a new broker for <500 events/min; Postgres partial index suffices (ADR-0014). |

## 6. Rollout & Verification Strategy
- **Feature flag & rollout:** Ship migration first; enable `WEBHOOK_RETRIES_ENABLED` in staging, then globally.
- **Vertical milestones (3–5 slicing boundaries):**
  1. Schema migration + store methods pass DB integration tests.
  2. Dispatcher retry/DLQ state machine verified with mock HTTP server tests.
  3. Admin replay endpoint + metrics verified end-to-end.

## 7. Open Questions (Must Be Empty Before Jira Slicing)
- [x] Retry `408`/`429` only; send other `4xx` to `dead_letter` -> **Decision:** Locked.
- [ ] [Open question] -> **Owner:** @[name] (Resolve before running `project-planning`)
```
