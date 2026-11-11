# Custom belief authoring

Scope: Authoring custom team and enterprise beliefs in `--additional-beliefs` directories.

## File naming and structure

- Place `.md` files in the configured `--additional-beliefs` folder or any nested subdirectory.
- The relative path becomes the `topic_id` (e.g. `security/secrets.md` → `topic_id: "security/secrets"`).
- To replace a built-in belief (e.g. `coding-standards`), match its filename (`coding-standards.md`).

## Fractal routing (Two-Tier discovery)

To prevent context bloat and wire budget blowout in large knowledge bases:
- **Root Topics (Top-Level)**: Only `.md` files at the root of the folder appear in the initial `cortex_bootstrap` router table.
- **Sub-Topics (Nested)**: Files in subdirectories are indexed in memory for fast retrieval (`cortex_get_belief`) and full-text search (`cortex_beliefs_search`), but are excluded from the root bootstrap table.
- **Domain Routers**: Place a top-level `.md` file matching the folder name (e.g. `security.md` for `security/`) to serve as a domain sub-router. If omitted, Cortex automatically synthesizes a domain router listing the sub-topics.

## YAML frontmatter specification

Declare topic metadata at the very top of the `.md` file using a YAML frontmatter block:

```markdown
---
trigger: When configuring AWS IAM roles, policies, or S3 bucket permissions
---
```

Supported fields:
- `trigger`: (Required for new topics) Concrete action phrase describing when an agent should load this belief.
- `title`: (Optional) Human-readable title. Defaults to the first `# Heading`.
- `id`: (Optional) Explicit topic ID. Defaults to relative filepath without `.md`.
