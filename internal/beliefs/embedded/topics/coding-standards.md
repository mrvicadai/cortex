Scope: Coding guidance for writing, refactoring, and fixing code.

## Loading

MUST read this whole document before writing, refactoring, or fixing code; loading a pointer alone is not sufficient.
Apply the mandatory [Open-Closed Principle (OCP) design checkpoint](#open-closed-composition) before editing behavior.
Treat Sound design seams, Useful API documentation, Preferred file size, and Preferred function length as advisory; apply requirements and prohibitions elsewhere.
Apply repository-specific conventions and graph-specific stricter caps before these general targets.
Check these size and complexity standards manually unless the repository provides enforcement; do not assume automation is installed.
Do not pad or compress code to satisfy metrics.

## Repository conventions

Repo conventions win over this guide; state a conflict instead of enforcing this guide.

## Proof guidance

Read testing guidance before claiming a fix is proven.

## Minimal implementation

Prefer the least code that produces the requested impact; simplify code that is materially larger than necessary.

## Readable and simple code

Use clear names, flat structure, and obvious flow.
Do not add unrequested flexibility, single-use abstractions, or impossible-case handling.

## Open-closed composition

MUST apply the Open-Closed Principle (OCP): keep stable core logic open for extension, closed for modification by composing small lambdas/functions, explicit policies, or typed configuration for real behavioral variation rather than growing central orchestration branches.
Before editing behavior, identify the stable core, the actual variation point, and the minimal composition that supplies the new behavior.
Use the smallest mechanism that fits the current requirement under [minimal implementation](#minimal-implementation) and [readable and simple code](#readable-and-simple-code); prefer typed configuration over stringly options or excessive feature flags.
Preserve injected dependencies instead of bypassing them by constructing global dependencies.
Keep rollout, source-selection, and rail-selection policy in composed functions or policies rather than embedding it in orchestration.
Correct defects and perform necessary refactors; this principle does not freeze existing code.

## Sibling precedent

Before copying a sibling, check matching scope, current naming, inert settings, and missing guards.
Flag newly-live or intentionally divergent settings.

## Surgical changes

Change only task-related lines, preserve local style, and mention unrelated dead code without deleting it.
Remove only unused code introduced by this change.

## Commit format

Commit each verified logical change immediately unless explicitly instructed not to commit; do not accumulate multiple changes across tasks. Keep each commit focused on one logical change and exclude unrelated work.
Use a single-line conventional commit with an allowed type and an action that says what changed and why.

## Callback locks

Before calling framework-owned callbacks, inspect the caller's lock lifetime; never re-enter a non-reentrant owner from its callback.

## Sound design seams

A deadlock invalidates an illegal call, not necessarily the design seam; repair the mechanism while retaining a sound design.

## Useful API documentation

Briefly document public APIs and complex private methods when their behavior is not obvious.

## Incidental comments

Do not add incidental inline comments unless requested; prefer self-documenting names, small functions, flat hierarchies, and imports over qualified names.

## Dependency approval

Do not introduce third-party dependencies without explicit user approval; report the need as a blocker.

## Risky behavior flags

Gate only risky, hard-to-reverse behavior.
A migration flag selects a resolution strategy, never a persisted per-entity routing mode.

## Preferred file size

Prefer cohesive files of 150–300 physical lines; shorter cohesive files are fine and need no padding.

## Preferred function length

Prefer functions of 20–50 physical lines; shorter cohesive functions are fine and need no padding.

## Cyclomatic complexity

MUST keep cyclomatic complexity strictly less than 8 per function (maximum 7).
