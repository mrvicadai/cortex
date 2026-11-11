Scope: Testing guidance for defining proof, running verification, and judging completion.

## Loading

Read this whole document before defining proof, running verification, or claiming tested or verified completion.

## Proof matched to claim

Define verifiable success criteria.
Validation needs invalid-input tests; bug fixes need reproductions; refactors need before-and-after passing tests.
Do not treat build or vet as runtime proof.

## Exact project checks

Run project typecheck, lint, and format where available, and run the specified exact test command rather than a guessed equivalent.

## Inversion

Watch the test fail without the fix and pass with it, even when confident.
Repair an inversion compile error instead of skipping inversion.

## Passing inversion

A passing inversion indicts the test.
Diagnose it, fix the test first, and otherwise qualify the claim as correct by construction rather than test-proven.

## Manufactured coverage

Do not quietly strengthen unrelated assertions until a test turns red; that manufactures apparent coverage instead of proving the claimed path.

## Divergence input

Name the input that diverges fixed and broken paths before asserting.
If the fixture cannot produce it, record that the test cannot catch the regression.

## Failure preconditions

Failure-path tests must actively break and assert the precondition before asserting the result.

## No-verdict runs

Treat a no-verdict run as a failure mode: check liveness, then use a short timeout that produces stacks; do not infer a failure from empty grep output or logs.

## Structural coverage

Check fields and branches actually exercised against real-client requests; green tests can be structurally blind.

## Read-guard regression

A read-path guard that only subtracts behavior can regress existing callers; find and mirror a relevant precedent before adding one.

## Behavior preservation

Verify behavior-preserving claims against real consumers and current-world cases; do not promote an untested hypothesis into a no-op assertion.

## Completion gate

Passing tests prove correctness; committing preserves proof.
MUST NOT claim a task or bug fix is tested, verified, or complete while modified files remain uncommitted in `git status`, unless the user explicitly instructed not to commit.
