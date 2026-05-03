---
name: siyuan-cli
description: Use when a task mentions siyuan, siyuan-cli, note-taking, thought-note workflows, or PKM and needs real CLI commands to search notes, inspect notebooks, read or update documents, export content, or automate SiYuan workflows instead of guessing HTTP API calls.
---

# siyuan-cli

## Overview

Use this skill when an agent should operate SiYuan through the real CLI in this repository. Prefer direct `siyuan-cli` commands over hand-written API requests for search, document edits, exports, notebook inspection, and other task-oriented note workflows.

## When to Use

- User mentions `siyuan`, `siyuan-cli`, note-taking, thought-note, or PKM workflows
- Need real `siyuan-cli` commands instead of guessed HTTP calls
- Need IDs or structured results from SiYuan for follow-up automation
- Need to read, create, update, move, export, or remove note content
- Need notebook, block, tag, file, snapshot, template, notification, SQL, or system operations

## Requirements

- Ensure `SIYUAN_BASE_URL` and `SIYUAN_TOKEN` are set

## Installation & Usage

### Pre-built Binary

Download from GitHub Releases or install via:

```bash
go install github.com/hcuong-me/siyuan-cli/cmd/siyuan@latest
```

### Homebrew (macOS)

```bash
brew install hcuong-me/tap/siyuan-cli
```

## Quick Reference

| Task | Command pattern |
| --- | --- |
| List documents in a notebook | `siyuan-cli doc list <notebook-id> --json` |
| Read a document | `siyuan-cli doc get <notebook-id> --path <path>` |
| Create a document | `siyuan-cli doc create <notebook-id> --path <path> --content-file <file>` |
| Update a document | `siyuan-cli doc update <notebook-id> --path <path> --content-file <file>` |
| Search blocks | `siyuan-cli search block <keyword> --json` |
| Search documents | `siyuan-cli search doc <keyword> --json` |
| Export Markdown | `siyuan-cli export markdown <doc-id> --output <file>` |
| List notebooks | `siyuan-cli notebook list --json` |
| Run SQL | `siyuan-cli sql query "<sql>" --json` |

## Command Groups

The CLI provides 13 command groups:

| Group | Description |
|-------|-------------|
| `system` | Version, time, boot status |
| `notebook` | Create, list, open, close, rename notebooks |
| `doc` | Create, read, update, remove, list documents |
| `block` | Get, update, insert, remove content blocks |
| `search` | Full-text search blocks and documents |
| `sql` | Execute read-only SQL queries |
| `tag` | Manage tags and tag-based queries |
| `attr` | Get/set block attributes (metadata) |
| `export` | Export to Markdown, HTML, PDF, DOCX |
| `template` | List, render templates |
| `snapshot` | Repository snapshots (backups) |
| `asset` | Upload and manage assets |
| `file` | Raw filesystem operations |

## Recommended Patterns

- **Verify command syntax first.** Before using any `siyuan-cli` command, run `siyuan-cli <group> <subcommand> -h` to confirm exact flag names, positional arguments, and available subcommands. Never guess flag names or argument positions — this is the root cause of most failed commands.
- **Prefer dedicated commands over SQL.** Always try `doc list`, `doc get`, `search`, `notebook list`, `block get`, etc. before reaching for `sql query`. SQL is a last resort when no dedicated command can solve the task.
- **Use `doc list` to discover documents.** `siyuan-cli doc list <notebook-id> --json` is the standard way to enumerate documents in a notebook. Do not use SQL queries or search as a substitute.
- **Prefer `--json`** when the result will be consumed by another command or parsed by an agent.
- **Use `--content-file`** for larger Markdown updates and `--content` for short inline edits.
- **Never use `SELECT *` in SQL queries.** Always specify the columns you need explicitly.
- Use `siyuan-cli <group> --help` when you need the full subcommand surface.

## Common Workflows

### List and Read Documents

```bash
# List all documents in a notebook
siyuan-cli doc list <notebook-id> --json

# Read a specific document by path
siyuan-cli doc get <notebook-id> --path "/path/to/doc"
```

### Search and Read

```bash
# Search for blocks containing a keyword
siyuan-cli search block "project roadmap" --json

# Search for documents containing a keyword
siyuan-cli search doc "project roadmap" --json

# Use the returned ID and path to read the document
siyuan-cli doc get <notebook-id> --path "/path/to/doc"
```

### Create and Update

```bash
# Create a new document
siyuan-cli doc create <notebook-id> --path "/Projects/MyProject" --content-file ./initial.md

# Update with new content
siyuan-cli doc update <notebook-id> --path "/Projects/MyProject" --content-file ./updated.md
```

### Export for Sharing

```bash
# Search for the document
siyuan-cli search doc "meeting notes" --json

# Export as Markdown
siyuan-cli export markdown <doc-id> --output ./meeting-notes.md
```

### Query with SQL (last resort)

```bash
siyuan-cli sql query "SELECT id, content, type FROM blocks WHERE type='h' LIMIT 10" --json
```

### Backup Before Changes

```bash
# Create backup
siyuan-cli snapshot create --memo "before-bulk-update"

# Make your changes
siyuan-cli doc update <notebook-id> --path "/Projects/MyProject" --content-file ./changes.md
```

## Examples

```bash
siyuan-cli doc list <notebook-id> --json
siyuan-cli search block "roadmap" --json
siyuan-cli doc get <notebook-id> --path "/Docs/Roadmap"
siyuan-cli doc update <notebook-id> --path "/Docs/Roadmap" --content-file ./draft.md
siyuan-cli export markdown <doc-id> --output ./export.md
```

## Safety Notes

- Destructive commands require explicit `--yes`; do not add it unless the task really intends mutation
- Double-check IDs, labels, and file paths before mutation
- If you need stable IDs from search, rerun with `--json`
- `doc create` and `doc update` can upload and rewrite Markdown image links automatically

## Common Mistakes

| Mistake | Fix |
| --- | --- |
| Guessing HTTP endpoints | Use the real `siyuan-cli` command family first |
| Forgetting env vars | Export `SIYUAN_BASE_URL` and `SIYUAN_TOKEN` before running commands |
| Parsing human text for IDs | Use `--json` |
| Using destructive commands casually | Omit `--yes` until the target is verified |
| Updating large Markdown inline | Prefer `--content-file` |
| Guessing flags without checking `-h` | Always run `<command> -h` first to confirm exact syntax |
| Using `sql query --statement "..."` | SQL is a positional argument: `sql query "..."` |
| Using `search --content "..."` | Search requires subcommand: `search block <keyword>` or `search doc <keyword>` |
| Using `export markdown --id ...` | doc-id is positional: `export markdown <doc-id>` |
| Using `doc get --id ...` or `doc update --id ...` | Both use `<notebook-id>` positional + `--path` flag |
| Trying to use `doc append` | Subcommand does not exist; use `doc update` to modify content |
| Using `SELECT *` in SQL queries | Always specify columns explicitly |
| Using SQL when a dedicated command exists | Try `doc list`, `doc get`, `search`, `notebook list` first; SQL is last resort |
