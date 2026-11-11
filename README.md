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

## Zero-Drift Project Setup (`AGENTS.md`)

Generate a compact, drift-free pointer for your project's `AGENTS.md`:

```bash
npx -y @mrvicadai/cortex snippet
```

This outputs an immutable, zero-drift directive:

```markdown
<!-- CORTEX_BELIEFS_START -->
## Agent Operating Beliefs & Context

**MANDATORY PROTOCOL**: At the start of every session, you MUST call 'beliefs_bootstrap' (or 'cortex_bootstrap') to load the core operating principles (ODAV, Least-to-Most, Project Memory) and the active task-triggered topics router.

Before executing any task phase (e.g. coding, testing, git stacking, designing, opening MRs), you MUST consult the relevant topic(s) discovered from the bootstrap router using 'beliefs_get(topics: ["<topic-id>"])' (or 'cortex_get_belief').
<!-- CORTEX_BELIEFS_END -->
```

### Why this is scalable:
1. **Dynamic Tool Schema Reflection**: In the MCP protocol, `cortex_get_belief` dynamically declares the active topics and triggers directly inside its tool definition on server handshake. The model receives active topic triggers on **Turn 0** with zero tokens wasted in `AGENTS.md`.
2. **Never Goes Stale**: Adding topics or configuring local overrides (`--additional-beliefs`) requires zero updates to `AGENTS.md`.

### In-Place Table Sync (`cortex sync`)
To sync any target `AGENTS.md` in-place, Cortex updates the block between `<!-- CORTEX_BELIEFS_START -->` and `<!-- CORTEX_BELIEFS_END -->` (or legacy markers) with the clean mandatory protocol:

```bash
npx -y @mrvicadai/cortex sync AGENTS.md
```

---

## Tools Interface

Cortex enforces strict zero-duplication: each capability exposes exactly one canonical `cortex_*` tool and one historical alias (`fieldnotes_*` or `beliefs_*`) for 100% backward compatibility:

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
| `cortex_bootstrap` | `beliefs_bootstrap` | Session bootstrap (principles + router) or query filter with SemVer & `origin` metadata |
| `cortex_get_belief` | `beliefs_get` | Batch retrieval of rule guidance (defaults to `topics: []` array) |
| `cortex_beliefs_search` | `beliefs_search` | Rule keyword search with optional inline markdown (`inline_content: true`) |

### 3. Cross-Cutting Search & Resources

* **`cortex_search`**: Multi-scope full-text search (`scope: "all" | "beliefs" | "fieldnotes"`).
* **Resource `cortex://beliefs/{topic}`**: Direct URI read access to full markdown rule documents.
* **Resource `cortex://context`**: Direct URI snapshot of active project facts & recent events.
* **Prompt `cortex_bootstrap`**: Pre-seeds agent context with active project facts + belief router.

---

## Custom Belief Overrides (`--additional-beliefs`)

Users and organizations can override standard packaged beliefs or define new team rules:

```bash
npx -y @mrvicadai/cortex --additional-beliefs=~/additional_beliefs
```
or via environment variable:
```bash
export CORTEX_ADDITIONAL_BELIEFS=~/additional_beliefs
```

### Features:
- **Hierarchical / Nested Scanning**: Subdirectories are recursively scanned (e.g. `security/secrets.md` $\rightarrow$ `topic_id: "security/secrets"`).
- **Overrides & Additions**: Any file matching a standard topic ID (e.g. `coding-standards.md`) overrides the packaged rule and reports `origin: "user_override"`. New files append as custom team topics.
- **YAML Frontmatter**: Optionally specify `id`, `title`, and dynamic `trigger` routing:
  ```markdown
  ---
  id: k8s-infra
  title: Kubernetes Infrastructure
  trigger: Before editing helm charts or deployment yamls
  ---
  # K8s Guidelines...
  ```

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
