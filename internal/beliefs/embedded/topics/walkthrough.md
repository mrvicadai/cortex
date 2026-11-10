# Walkthroughs

Scope: Standalone walkthroughs of completed infrastructure or configuration changes, major code changes, or whole multi-MR projects for someone who was not in the room.

## Trigger and artifact

Use this guidance when the user types `[WALKTHROUGH]`, asks to be walked through a change, onboarded to what was built, or given the whole picture of a project.
This is a separate deliverable, not project state.
It authorizes exactly one standalone Markdown file in the project; it does not authorize a second one or make the document maintained.
Follow [project-state policy](../operating/project-state.md) for persistence.

## Evidence and authority

Read real configuration; never write from recall.
Open files and quote actual hostnames, bucket names, TTLs, and policy JSON.
Name the spec of record and defer to it.
The header must carry the date, head SHAs, a line that MR descriptions and tickets govern, and a point-in-time, not-maintained warning.
Cite `file:line`; excerpt sparingly, at most 10 lines.
State the live state of every flag and rollout, not merely that it exists.

## Orientation and structure

For a multi-MR project, lead with landing order.
For code, answer where execution enters before modules.
For a project, answer what shipped versus planned before any diagram.
Use a framing paragraph, inventory table, one section per part, a "what this does NOT do" section, and a quiz; choose inventory, diagrams, and exclusions by subject.
Explain traps and non-obvious choices, not just the happy path.
Assume an audience new to the stack with limited attention: frame in one plain-English paragraph; table first, then stand-alone short sections; use tables for more than two dimensions; bold each section's point; call out deliberate divergences.

## Diagrams

Pick the Mermaid diagram type by question: `flowchart`, `sequenceDiagram`, `stateDiagram-v2`, or `erDiagram`.
Use no styling or subgraph nesting beyond one level; quote every label; use one idea per diagram; use dotted edges for built but not live behavior.

## Quiz and handover

Target approximately 80% coverage of essentials in the quiz, with one fact per question.
Collapse answers with `<details><summary>Answer</summary>…</details>`, order by document order, and state why in each one- or two-sentence answer.
Before handover, count questions, `<details>`, and Mermaid fences; confirm fences balance and stray `<active>` and `<actve>` tags are zero.
State the counts.
