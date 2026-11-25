# Knowledge-base governance

Scope: Rules for maintaining `internal/beliefs/embedded/`.
Read this file before editing embedded beliefs, including during `[SELF_IMPROVE]`.

## Content and size

Keep only rules and load-bearing facts that would change an agent's actions.
Keep imperatives, detection mechanisms, concrete anchors, refuted hypotheses in imperative form, and pointers to the owning file; cut one-off history, incident cost, occurrence counts, reasoning, and summaries of another file.
Permit narrative only in imperative form.
Never restate another file's rule; link it instead.
Make citations identify the intended source and do not imply support the source does not provide.
Preserve source meaning during relocation or refactoring; do not manufacture claims or relationships to improve counts.
Offline link validation does not establish that external links are available or that a cited source supports a claim.
Never write meta-commentary about this base's past; write the imperative it implies instead.
Prefer an enforcing mechanism; when none enforces a constraint, label the rule aspirational rather than implying enforcement.
Replace a rule requiring judgement when judgement is compromised with one that removes the choice.
Keep maintained rule Markdown at or below 120 physical lines; split over-limit rules rather than compressing them.
Project state uses the caps and update rules in [project-state policy](../foundational/project-state.md).
Before and after editing Markdown, record `wc -c` and `wc -l`; reflow without content removal is not pruning.
Write each Markdown prose sentence on its own physical line; preserve structural formatting for code blocks, tables, and other non-prose content.
On every rule revisit, delete stale, duplicated, or decision-neutral text before adding anything; project-state history follows its own policy.
If no rule text qualifies, state that in the commit body.
Measure per heading before cutting and split along read moments: if a reader who needs A would not need B immediately, make two files.
State a leaf file's scope in one line at the top and include a decay warning for point-in-time facts: `Verified <date>; verify any file:line before asserting it`.

## Routing and precedence

Topic routing is generated dynamically by `cortex_bootstrap` and `cortex_get_belief`; never duplicate a topic's rules in the router trigger or `AGENTS.md`.
Repair a skipped rule's routing trigger or `related_topics` rather than duplicating the rule.
A target repository's own conventions, docs, and CI outrank these files, which outrank general judgement.
Maintain topic guidance in Markdown without duplicate prose; external instruction files retain only the bootstrap and approved tool-specific anchors.
After changes to `internal/beliefs/embedded/`, run `go test -v -count=1 ./internal/beliefs/...`.

## Self-improvement

On every user correction, harvest only the rule that would have prevented it; for `[SELF_IMPROVE]`, also mine `cortex_context` and `cortex_find(kind: "event")` for broken rules, re-derived facts, refuted hypotheses, and costly environment quirks.
Search existing rules via `cortex_beliefs_search` or `cortex_search(scope: "beliefs")` before writing each candidate; a hit is a routing or wording defect, so fix the trigger or wording and write nothing new.
Convert each surviving lesson into one imperative rule or load-bearing fact, place it with its owner, and create a routed file only when it has no owner.
On the third failure of the same rule, replace it with a mechanism or choice-free instruction; do not reword it.
Prune and measure in the same self-improvement pass, split above 120 lines, commit `docs: <what the rules now say>`, record via `cortex_event_append`, and return file → rule, one line each, with no narrative.

## Git safety

Before squashing, confirm every commit is yours, then confirm `git rev-parse HEAD^{tree}` is identical before and after; never silently absorb a peer commit.
