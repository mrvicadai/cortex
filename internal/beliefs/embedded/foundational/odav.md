# Observe, decide, act, verify (ODAV)

Scope: Evidence-driven execution and iterative improvement in every agent session.
Apply this methodology as agent policy; no automated enforcement is provided.

## Observe

Inspect the current state, applicable instructions, and relevant evidence before choosing an action.
Separate verified facts from assumptions and identify what remains unknown.

## Decide

Choose the next bounded action using the observed state and the user's requested outcome.
Define its success check and stopping condition before acting.
Use [least-to-most](least-to-most.md) to order dependent subproblems.

## Act

Perform the chosen action within the authorized scope.
Keep the change small enough to verify independently.

## Verify, commit checkpoint, and repeat

Check the actual result against the success criteria under [testing guidance](testing.md) when applicable.
Treat an attempted action, successful tool call, or unchanged assumption as insufficient proof of the requested outcome.

**MANDATORY COMMIT CHECKPOINT**: You MUST commit verified changes (`git add` + `git commit`) as a focused checkpoint before deciding the next action or concluding the turn, unless explicitly instructed not to commit. A verified change that remains uncommitted is an incomplete action.

Use the verified result to choose the next action; revise the plan when evidence changes rather than repeating it unchanged.
Persist applicable verified facts and completed work under [project-memory policy](project-state.md).
Before iterative improvement, define a target and iteration budget; stop when the target is met, improvement falls below the stated threshold, or the budget is exhausted.
For performance work, compare each iteration against a reproducible baseline and report unmet targets.
Keep trivial work to one lightweight cycle rather than manufacturing a planning ritual.
