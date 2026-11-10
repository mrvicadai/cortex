Scope: Verify claimed defects and completed dependencies before committing work.

## Verify claims before work

Verify a claimed defect or completed dependency before starting implementation.
Tickets, code, and merged Terraform are not systems of record.

## Query the system of record

Use one live query before implementation: logs for log claims, error/APM for errors, metrics backend for metrics, endpoint request for endpoint behavior, DB timing for slow queries, provider state for completed-resource claims.

## Nonbehavioral work

Pure refactors, typed-API work, and test-only work skip the premise gate.

## Act on the result

Stop, report, and re-scope when the premise is false; do not implement a no-op for a ticket.
Preserve a holding query as MR evidence when it is true.

## Merge and convergence

Done means code merged, not external convergence.
If external state cannot be read, disclose the access hazard in the MR description.

## Advisory premise staleness

Prioritize tickets by premise staleness and use telemetry, endpoints, CI, or forge APIs before spending a user round trip.

## Existing telemetry

Read existing instrumentation around the suspect path before proposing more; grep known log constants first.

## Trace across services

Trace a request across all services by trace ID rather than searching one service's errors.

## Record negative findings

Record completed negative-premise checks in project memory under [project-state policy](project-state.md).
