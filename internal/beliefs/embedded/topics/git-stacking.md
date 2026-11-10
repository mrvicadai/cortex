# Git stacking

Scope: Worktree creation, branch topology, conflict resolution, stack synchronization, and merge-request splitting.

## Required practices

Create new worktrees by default at dedicated directories (e.g., `~/worktrees/<task-name>`) rather than temporary scratch paths.
Resolve conflicts by merging origin main into the branch and pushing normally; port local work onto upstream shape and never revert upstream.
Commit a merge immediately before tests or generation, verify a two-parent commit, and put later formatting or generation in a separate follow-up commit.
Fix generation or formatting drift with the full documented generation command; individual formatting or dependency tools are insufficient.
Find a stack parent, create dependent or independent branches with the documented topology commands, and target the parent rather than main.
Synchronize the stack rather than cherry-picking or hand-rolling rebase.
Confirm parent remote state because sync reparents landed parents automatically.
Resolve stack conflicts in place and continue; do not abort into a worktree, and inspect rewritten add/add conflicts for duplicated blocks.
Validate changed paths and descendants that assert on them; a clean topology operation is not a green build.
Open against the parent, probe and record opening mechanisms, state pending MR status plainly when needed, verify draft state through metadata, and never merge, ship, or force-push.
Classify oversized diffs as prep, structural, or tests.
Land mergeable prep first, keep switch-dependent work structural, and prove the combined trees match the original.
Force-with-lease rewriting needs explicit user authorization and confirmation that the remote tip equals the base.

## Prohibited topology commands

Use only verified topology commands.
Do not ship, delete, force-push, or use browser-opening propose automation.

## Advisory: ticket size

Fold trivial refactors into feature work and split genuinely independent concerns when an MR exceeds the rough size trigger.
