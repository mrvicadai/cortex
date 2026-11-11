# Least-to-most problem solving

Scope: Solving complex tasks through small, dependency-ordered, verified subproblems in every agent session.
Apply this strategy as agent policy; no automated enforcement is provided.

## Decompose

Identify the requested outcome, constraints, and unknowns before tackling the full problem.
Break nontrivial work into the smallest useful subproblems with explicit dependencies and verifiable outcomes.
Order prerequisites before dependent work; prefer a simple foundational slice over a speculative end-to-end change.
Keep trivial tasks direct rather than inventing unnecessary subproblems.

## Solve and build

Solve the simplest ready subproblem using the [ODAV loop](odav.md).
Verify its result before using it as a premise or building a dependent slice on it.
**MANDATORY SLICE COMMIT**: You MUST commit each independently verified slice as a focused checkpoint (`git add` + `git commit -m "<type>: <action>"`) and record the milestone outcome via `cortex_event_append` under [project-state](project-state.md) before beginning the next dependent subproblem, unless explicitly told not to commit. Do NOT accumulate multiple verified slices into a single end-of-task commit.
Carry verified findings and reusable intermediate results into the next, more complex subproblem instead of re-deriving them.
Revise the remaining decomposition when new evidence invalidates a dependency or assumption.
Stop and resolve a failed prerequisite rather than building on an unverified result.

## Integrate

Verify the combined result against the original outcome and constraints; individually passing slices do not establish integration correctness.
**MANDATORY COMPLETION GATE**: Confirm `git status` is clean and all completed progress milestones are recorded via `cortex_event_append`. You MUST NOT claim the task is complete or hand off to the user with uncommitted changes unless the user explicitly instructed not to commit.
Report remaining unknowns and unmet criteria without claiming the whole task is complete.
