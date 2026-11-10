# Project design

Scope: Authoring and rendering project designs, RFCs, and technical specifications.

## Constraints and design lenses

Before any design, write what must not happen.
Anchor each entry to a verified fact or mark it inference.
Judge a candidate design by how many entries it structurally prevents — not by its elegance.
Name the consequence of every entry.
A pitfall with no consequence is a preference.
Bind each entry to the path where its mechanism was observed.
Do not generalize a guard seen on one code path into a guarantee the system does not offer.
Pair the negative list with what must be PRESENT — the few conditions whose absence no design choice recovers.
Over about seven, it is a task list, not a set of success conditions.
Carry the list into each candidate design and cite the entries it answers.
Lenses must disagree structurally, not stylistically.
Cover artifact, sequence, and boundary; add one lens arguing the opposite state/storage choice.
Argue each lens as well as it can honestly be argued and name its weaknesses.
Forbid hedging toward a compromise.
Require a REJECTS section naming what the design refuses and what that refusal costs.
Convergence from different premises outranks a majority.
Two designs agreeing because they share an assumption is one data point, not two.
Name what survives from a losing design before discarding it.
Verify each design's load-bearing positives yourself before synthesis.

## Design content

A technical design document and a delivery plan are two artifacts.
Do not merge them.
Keep mechanism in the design, including cutover and rollback behavior.
Schedules, dates, estimates, milestones, staffing, ticket breakdowns, delivery sequencing, and management status belong to the plan.
Give an RFC or design subgraph only the sections needed to judge its mechanism: decision summary; scope and constraints; architecture and runtime behavior; applicable state, data, and contracts; failure, recovery, and reversibility; rejected alternatives; and materially design-changing open questions.
Omit irrelevant or empty sections.
Render the approved implementation specification under [Technical specification rendering](#technical-specification-rendering) instead.
Treat spikes and discovery inventories as evidence, not RFC sections.
Cite them and carry forward only facts that constrain the selected design.
Rejected alternatives are design content and stay.
How the document was produced is not.
State the largest work item up front and say plainly when it is net-new construction rather than a migration or repointing step.
A defect found during design is not thereby a design topic.
Check who owns the subsystem before writing it up.
Design against a dependency's published interface, never its internals, even when you can read their code.
Before proposing a new component, enumerate what already exists at that seam.
Search for the controller, route, entity, or table you are about to introduce.
Use the target's real constructors and helpers, not the library's.
Cite the wrapper found in a sibling module.
When an open question gets answered, delete the scaffolding built to manage the uncertainty.
Anything genuinely out of scope gets one row saying so, and no analysis.

## Diagrams and models

Use Mermaid as the specification when structure or temporal behavior would otherwise require verbose prose: `flowchart` for components and data flow, `sequenceDiagram` for runtime interactions, `stateDiagram-v2` for valid transitions, and `erDiagram` for data relationships.
Keep one idea per diagram; prose records decisions and exceptions, never a narration of its arrows.
Validate each diagram with its target renderer before delivery; renderer grammar varies, so generated diagrams must avoid `sequenceDiagram` `Note over` syntax and put explanatory constraints in adjacent prose.
In an RFC/design subgraph or rendered technical specification, render data models as expressive TypeScript types, interfaces, and discriminated unions; use JSON Schema, Protobuf, or SQL DDL only when a target or external contract requires that artifact, and do not manually maintain an equivalent second schema.
Encode every sequence as Mermaid or precise pseudocode; use exact terminology and omit narrative prose wherever a type, diagram, or code block conveys the design.

## Technical specification rendering

When rendering an approved design subgraph as the final human-readable implementation artifact, start with `# Technical Specification: [System Name]` and use exactly these H2 sections: `System Topology & Component Boundaries`; `Data Models & Database Schemas`; `API Contracts (Endpoints, Payload Formats, Status Codes)`; `State Transitions & Processing Logic`; `Failure Modes, Edge Cases & Circuit Breakers`.
In a generated technical specification, omit greetings, preambles, sign-offs, historical context, project background, business motivation, external references, timelines, estimates, staffing, phases, roadmaps, and project-management metrics.

## Audience

Name ONE audience before writing.
A second audience that genuinely needs something requires a second document.
A design doc is standalone: no session citation IDs, cross-document pointers, process narrative, revision changelog, or internal source anchors.
State evidence as the constraint, decision, uncertainty, or externally meaningful reference; keep audit anchors in the knowledge graph or a separate evidence artifact.
Reader-facing RFC, technical-design, and technical-specification Markdown must not expose internal `path:line` citations.
Use internal anchors while researching, then render only the supported conclusion, constraint, uncertainty, or separate evidence reference.
Write for a busy engineer who will skim: lead each section with its conclusion, use tables for multidimensional information, and cut prose that takes more than two lines to reach a point.
