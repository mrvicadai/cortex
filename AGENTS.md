<!-- CORTEX_BELIEFS_START -->
## Agent Operating Beliefs & Context

**MANDATORY PROTOCOL**: At the start of every session, you MUST call 'beliefs_bootstrap' (or 'cortex_bootstrap') to load the core operating principles (ODAV, Least-to-Most, Project Memory) and the active task-triggered topics router.

Before executing any task phase (e.g. coding, testing, git stacking, designing, opening MRs), you MUST consult the relevant topic(s) discovered from the bootstrap router using 'beliefs_get(topics: ["<topic-id>"])' (or 'cortex_get_belief').

**DEFAULT TO COMMIT**: Every verified code change or completed slice MUST be committed immediately as a git checkpoint before proceeding or claiming completion, unless explicitly instructed otherwise.
<!-- CORTEX_BELIEFS_END -->

