<!-- CORTEX_BELIEFS_START -->
## Agent Operating Beliefs & Context

**MANDATORY PROTOCOL**: At the start of every session, you MUST call 'cortex_bootstrap({})' (with zero arguments) to load the core operating principles (ODAV, Least-to-Most, Project Memory) and the active task-triggered topics router.

Before executing any task phase (e.g. coding, testing, git stacking, designing, opening MRs), you MUST consult the relevant topic(s) discovered from the bootstrap router using 'cortex_get_belief(topics: ["<topic-id>"])'.

**DEFAULT TO COMMIT**: Every verified code change or completed slice MUST be committed immediately as a git checkpoint before proceeding or claiming completion, unless explicitly instructed otherwise.

**RECORD PROGRESS MILESTONES**: After completing ANY major action — including investigations that discover a root cause, diagnostic findings, verified test runs, or code commits — you MUST immediately append a milestone event using 'cortex_event_append' with supporting evidence. Do not wait for user prompting or session end; investigations without code changes still require event recording.
<!-- CORTEX_BELIEFS_END -->
