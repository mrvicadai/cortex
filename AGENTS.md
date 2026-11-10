<!-- CORTEX_START -->
## Agent Operating Context (Cortex)
This repository uses Cortex for behavioral guidance and episodic memory:
- **Foundations**: Apply ODAV (Observe, Decide, Act, Verify), Least-to-Most subproblem ordering, and active Project Memory.
- **Guidance on Demand**: The active topic router is declared directly in your 'cortex_get_belief' (or 'beliefs_get') tool definition. Consult relevant topic(s) before writing code, testing, git stacking, designing, or opening MRs. Call 'cortex_list_beliefs' to inspect all topics.
- **Episodic Memory**: At session start, call 'cortex_context' to confirm project identity. Persist facts with 'cortex_fact_put' and milestones with 'cortex_event_append'.
<!-- CORTEX_END -->

