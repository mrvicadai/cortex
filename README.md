# `@mrvicadai/cortex`

> **Unified Agent Context Architecture**  
> Consolidating normative operating beliefs and persistent episodic memory for AI coding agents into a single, high-performance Go MCP server with zero-install `npx` delivery.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## Overview

Agentic workflows require two fundamentally different types of context to operate effectively:

1. **Normative Context (Beliefs):** Cross-cutting principles, coding standards, heuristics, and behavioral guidelines (*"how the agent should think and behave"*). Packaged as embedded SemVer markdown rules with user-override capability (`--additional-beliefs`), discovered progressively to prevent context bloat.
2. **Episodic Context (Fieldnotes):** Empirical project facts, decisions, blockers, and milestone progress updates (*"what is true about this specific project right now"*). Managed in persistent pure-Go SQLite with CAS optimistic concurrency (`expected_revision`), idempotency receipts (`operation_id`), structured evidence pointers, and strict 8 KiB context budgeting.

**Cortex** unifies both systems into a single MCP package, eliminating dual-server configuration, fragmented queries, and tool confusion.

---

## Quickstart & Client Configuration

Zero-install via `npx` (requires Node.js 18+ and Go 1.23+ if building locally):

### 1. Claude Desktop
Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "cortex": {
      "command": "npx",
      "args": ["-y", "@mrvicadai/cortex"]
    }
  }
}
```

### 2. Cursor
Add to **Settings $\rightarrow$ Features $\rightarrow$ MCP Servers**:
* **Name:** `cortex`
* **Type:** `command`
* **Command:** `npx -y @mrvicadai/cortex`

### 3. Antigravity / AGY (with Eager Tool Registration)
```json
{
  "mcpServers": {
    "cortex": {
      "command": "npx",
      "args": ["-y", "@mrvicadai/cortex"],
      "forceAllToolsEager": true
    }
  }
}
```

---

## Zero-Turn Project Router (`AGENTS.md`)

Generate a ready-to-paste router block for your project's `AGENTS.md`:

```bash
npx -y @mrvicadai/cortex snippet
```

This injects the foundational principles (ODAV, Least-to-Most, Project-State) and the 10-topic router table directly into the repo's instructions:

```markdown
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
<!-- CORTEX_BELIEFS_END -->
```

---

## Tools Interface

Cortex exposes primary `cortex_*` tools alongside dual-registered historical aliases (`fieldnotes_*`, `beliefs_*`) for 100% backward compatibility:

### 1. Episodic Context Tools (Fieldnotes Engine)

| Primary Tool | Historical Alias | Purpose | Mutating / CAS |
| :--- | :--- | :--- | :--- |
| `cortex_context` | `fieldnotes_context` | Load `@context` guidance & recent events within 8 KiB budget | Read-only |
| `cortex_fact_put` | `fieldnotes_fact_put` | Create or CAS-replace a persistent project fact | Yes (CAS + `operation_id`) |
| `cortex_fact_retire`| `fieldnotes_fact_retire`| CAS-retire a fact, preserving an opaque tombstone | Yes (CAS + `operation_id`) |
| `cortex_event_append`| `fieldnotes_event_append`| Append immutable outcome to project ledger with evidence | Yes (`operation_id`) |
| `cortex_find` | `fieldnotes_find` | Search/browse references across facts & events | Read-only |
| `cortex_read` | `fieldnotes_read` | Fetch 1–10 unchunked complete records with tombstone status | Read-only |

### 2. Normative Belief Tools (Belief System Engine)

| Primary Tool | Historical Alias | Purpose |
| :--- | :--- | :--- |
| `cortex_list_beliefs` | `beliefs_bootstrap`, `cortex_beliefs_bootstrap` | Session bootstrap (principles + router) or query filter with SemVer & `origin` metadata |
| `cortex_get_belief` | `beliefs_get`, `cortex_beliefs_get` | Batch retrieval of rule guidance (defaults to `topics: []` array) |
| `cortex_beliefs_search` | `beliefs_search` | Rule keyword search with optional inline markdown (`inline_content: true`) |

### 3. Cross-Cutting Search & Resources

* **`cortex_search`**: Multi-scope full-text search (`scope: "all" | "beliefs" | "fieldnotes"`).
* **Resource `cortex://beliefs/{topic}`**: Direct URI read access to full markdown rule documents.
* **Resource `cortex://context`**: Direct URI snapshot of active project facts & recent events.
* **Prompt `cortex_bootstrap`**: Pre-seeds agent context with active project facts + belief router.

---

## Custom Belief Overrides (`--additional-beliefs`)

Users can override standard packaged beliefs or define new team rules:

```bash
npx -y @mrvicadai/cortex --additional-beliefs=~/additional_beliefs
```
or via environment variable:
```bash
export CORTEX_ADDITIONAL_BELIEFS=~/additional_beliefs
```

Any `.md` file matching a standard topic ID (e.g. `coding-standards.md`) will override the packaged rule and report `origin: "user_override"`. New files will append as custom team topics.

---

## Database Storage & Compatibility

Cortex stores episodic data in pure-Go SQLite (`modernc.org/sqlite` in WAL mode):
- Default location: `.cortex/cortex.sqlite` (or `~/.cortex/cortex.sqlite`).
- If an existing `.fieldnotes/fieldnotes.sqlite` file is present in the workspace, Cortex automatically connects to it with zero configuration!
- Fully binary compatible with existing Fieldnotes database files (`application_id = 0x464e4f54`).

---

## Development & Testing

```bash
# Clone and enter repo
cd ~/repos/cortex

# Run all test suites
npm test # or go test -v ./...

# Build binary
npm run build # or go build -o bin/cortex ./cmd/cortex

# Start via CLI shim
node ./bin/cortex.js --help
```

---

## License

MIT © [Aaron Dai](https://github.com/mrvicadai)
