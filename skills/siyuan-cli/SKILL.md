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
- Need to read, create, update, append, move, export, or remove note content
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
brew tap hcuong-me/siyuan-cli
brew install siyuan-cli
```

## Quick Reference

| Task | Command pattern |
| --- | --- |
| Search notes | `siyuan-cli search --content <text> --json` |
| Read a document | `siyuan-cli doc get --id <document-id>` |
| Create a document | `siyuan-cli doc create --notebook <id> --path <path> --content-file <file>` |
| Update a document | `siyuan-cli doc update --id <document-id> --content-file <file>` |
| Append to a document | `siyuan-cli doc append --id <document-id> --content <text>` |
| Export Markdown | `siyuan-cli export markdown --id <document-id> --output <file>` |
| List notebooks | `siyuan-cli notebook list --json` |
| Run SQL | `siyuan-cli sql query --statement <sql> --json` |

## Command Groups

The CLI provides 13 command groups:

| Group | Description |
|-------|-------------|
| `system` | Version, time, boot status |
| `notebook` | Create, list, open, close, rename notebooks |
| `doc` | Create, read, update, move documents |
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

- Prefer `--json` when the result will be consumed by another command or parsed by an agent
- Use search first, then feed returned IDs into `doc`, `export`, `block`, or `attr` commands
- Use `--content-file` for larger Markdown updates and `--content` for short inline edits
- Use `siyuan-cli <group> --help` when you need the exact subcommand surface

## Common Workflows

### Search and Read

Search for content, then read the document:

```bash
# Search for documents
siyuan-cli search --content "project roadmap" --json

# Use the returned ID to read the document
siyuan-cli doc get --id 20260316120000-abc123
```

### Create and Update

Create a new document, then update it:

```bash
# Create a new document
siyuan-cli doc create --notebook 20260316120000-notebook1 --path "/Projects/MyProject" --content-file ./initial.md

# Update with new content
siyuan-cli doc update --id 20260316120000-abc123 --content-file ./updated.md
```

### Export for Sharing

Search and export to shareable format:

```bash
# Find the document
siyuan-cli search --content "meeting notes" --json

# Export as PDF
siyuan-cli export pdf --id 20260316120000-abc123 --output ./meeting-notes.pdf
```

### Query with SQL

Direct database queries for advanced filtering:

```bash
siyuan-cli sql query --statement "SELECT id, content FROM blocks WHERE type='p' LIMIT 10" --json
```

### Backup Before Changes

Create a snapshot before making bulk changes:

```bash
# Create backup
siyuan-cli snapshot create --memo "before-bulk-update"

# Make your changes
siyuan-cli doc update --id 20260316120000-abc123 --content-file ./changes.md
```

## Examples

```bash
siyuan-cli search --content "roadmap" --json
siyuan-cli doc get --id 20260316120000-abc123
siyuan-cli doc update --id 20260316120000-abc123 --content-file ./draft.md
siyuan-cli export markdown --id 20260316120000-abc123 --output ./export.md
```

## Safety Notes

- Destructive commands require explicit `--yes`; do not add it unless the task really intends mutation
- Double-check IDs, labels, and file paths before mutation
- If you need stable IDs from search, rerun with `--json`
- `doc create`, `doc update`, and `doc append` can upload and rewrite Markdown image links automatically

## Common Mistakes

| Mistake | Fix |
| --- | --- |
| Guessing HTTP endpoints | Use the real `siyuan-cli` command family first |
| Forgetting env vars | Export `SIYUAN_BASE_URL` and `SIYUAN_TOKEN` before running commands |
| Parsing human text for IDs | Use `--json` |
| Using destructive commands casually | Omit `--yes` until the target is verified |
| Updating large Markdown inline | Prefer `--content-file` |
