# Feasibility and dependency audit (Stage 1)

Scope: Pre-RFC read-only codebase, contract, and external dependency verification before `rfc-authoring`.

## Cortex memory first (anti-bloat)

- Query `cortex_find(kind: "fact", query: "<domain>")` and `cortex_read` matching IDs first to reuse already-verified system invariants, schemas, and service contracts.
- Persist newly discovered **durable** dependency or contract constraints via `cortex_fact_put` (`topics` + `evidence`); never dump raw files or unverified guesses into Facts.
- Record the Stage 1 audit summary via `cortex_event_append` when requesting Gate 1 clearance.

## Five required verification areas

| Area | Verification Requirement |
| :--- | :--- |
| **External APIs & SDKs** | Verify exact endpoints, HTTP methods, payload schemas, and SDK versions against official docs, OpenAPI specs, or internal clients. |
| **Data Contracts** | Verify exact payload fields, types, nullability constraints, and event delivery guarantees. |
| **Auth & Permissions** | Verify required OAuth scopes, API key tier entitlements, IAM roles, and tenant isolation policies. |
| **Quotas & Limits** | Verify rate limits, payload size caps, pagination mechanisms (cursor vs. offset), and SLA expectations. |
| **Gaps & Blockers** | Surface any missing component, client method, or config before locking in an unbuildable design. |

## Required output: Dependency & Capability Matrix

Produce this matrix before requesting Gate 1 clearance:

```markdown
| Dependency / System | Required Capability or Data | Status (Verified / Partial / Missing / Unverified) | Evidence (file path, spec, or doc URL) | Constraints (rate limits, auth, pagination, deprecated) |
| :--- | :--- | :--- | :--- | :--- |
| [System Name] | [Required interface/data] | [Status] | [File/URL] | [Constraints] |
```

## Gate 1 clearance

- Exhaust static proof (repo code, tests, schemas, `cortex_find`, and official docs) before marking any dependency `Unverified`.
- If a critical dependency is `Missing` or `Unverified`, pivot back to Stage 0 (`autonomous-design`) via *Via Negativa* or bound the unknown with a safe fallback; escalate *only* on true external dead-ends (missing credentials/permissions, private undocumented APIs, or contradictory goals).
- Pause for human Gate 1 approval before Stage 2 (`rfc-authoring`) only when `hitl-design-gates` is active.
