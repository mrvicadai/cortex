Scope: Build and verify Datadog telemetry queries, source selection, facets, and evidence interpretation.

## Origin telemetry for endpoint error breakdowns
**Required**
For an endpoint ERROR BREAKDOWN, build on origin/BFF telemetry, never the CDN edge.
Edge logs can count 5xx but cannot classify causes; correlate back to APM by resource name plus timestamp.

## Edge-terminated 504 blind spot
**Required**
An edge-terminated 504 produces ZERO origin spans.
Keep exactly one edge series as a "did the request even reach us" counter and label that blind spot on the widget itself.

## Discover Datadog skills first
**Required**
Discover skills first, in parallel: call `load_datadog_skill` with a domain and `list_datadog_skills` with fuzzy keywords.
Load matches plus what they point to.
`datadog/logs` requires `datadog/ddsql` loaded first.

## Aggregate telemetry server-side
**Required**
For generic server-side MCP aggregation guidance, follow the Context hygiene guide.
In Datadog, use `aggregate_events`, `aggregate_spans`, or `aggregate_rum_events` for counts, rates, and distributions; reserve `search_*` for raw-event inspection.

## Restrict telemetry fields
**Prohibited**
Follow the Context hygiene guide for MCP field limits.
Never pass `extra_fields: ["*"]`; name only the required Datadog attributes.

## Verify telemetry facets
**Required**
Never invent a group-by column or facet.
Source it from skill discovery, evidence linked by project facts under [project-state policy](../operating/project-state.md), or `search_*` with named `extra_fields`.
An empty group column means stop and look up the name.

## Use qualified Datadog tool names
Follow the Context hygiene guide for fully-qualified ToolSearch naming.
Datadog tool names use the `mcp__datadog-mcp__` prefix, such as `mcp__datadog-mcp__search_datadog_logs`.

## Quantify logging premises
**Required**
For a logging ticket, premise-check with the percentage of log lines carrying field X, including the denominator.

## Group outer log messages
**Required**
Group on the OUTER `message`, not `@error.message`.
Read one raw log line before ranking a "new" group.

## Rank telemetry daily
**Required**
Rank on daily buckets, never a short window.

## Count traces not logs
**Required**
Count trace ID cardinality, not log lines, before calling a count N incidents.

## Avoid false-negative shell chains
**Prohibited**
Never conclude from `cmd && echo YES || echo NO`.
Test existence first with `git cat-file -e <sha>^{commit}`, then compare.

## Verify version tag provenance
**Prohibited**
A Datadog `version` tag is not necessarily a commit in the current repo.
After `git fetch`, an unresolvable SHA may be another repo or build hash; check deployment manifests or upstream builds.
