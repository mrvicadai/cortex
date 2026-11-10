<!-- CORTEX_BELIEFS_START -->
## Agent Operating Beliefs

You operate under core operating principles and progressive topic guidance:
- **ODAV**: Observe, Decide, Act, Verify. Inspect state and verified facts before acting; test each change.
- **Least-to-Most**: Decompose complex tasks into small, dependency-ordered, verified subproblems.
- **Project Memory**: Proactively record verified milestones, facts, and validation results.

### Task-Triggered Topics Router
Before planning or executing any of the following tasks, call `beliefs_get(topics: ["<id>"])`:

| Topic ID | Topic Name | When to Call |
| :--- | :--- | :--- |
| `coding-standards` | Coding Standards | Before writing, refactoring, or fixing code |
| `testing` | Testing & Verification | Before defining proof, running verification, or claiming completion |
| `mr-rules` | Merge Request / Pull Request Rules | Before opening or updating an MR or PR |
| `git-stacking` | Git Stacking & Topology | Creating worktrees, branch topology, or stacking |
| `project-design` | Project Design & Technical Specifications | Designing a project, writing an RFC, or technical specification |
| `walkthrough` | Project Walkthroughs | Authoring a whole-project walkthrough or onboarding explainer |
| `datadog-query-craft` | Datadog Query Craft & Telemetry | Querying Datadog telemetry, APM traces, error breakdowns, or logs |
| `premise-check` | Premise Verification | Checking a claimed defect or dependency premise before committing work |
| `context-hygiene` | Context Hygiene & Token Conservation | Reducing token overhead in MCP queries, tool output, or shell commands |
| `governance` | Knowledge-Base Governance | Maintaining belief rules or performing self-improvement |

*Usage*: Call `beliefs_get(topics: ["<id>"])` when entering a task phase. Use `beliefs_search(query: "<text>")` to search rules, or `beliefs_bootstrap` for full operating principles.
<!-- CORTEX_BELIEFS_END -->

