# siyuan-cli Command to API Mapping

This document maps CLI commands to their corresponding SiYuan API endpoints.

---

## System

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan-cli system version` | `POST /api/system/version` | Get SiYuan kernel version |
| `siyuan-cli system time` | `POST /api/system/currentTime` | Get current system time |
| `siyuan-cli system boot-progress` | `POST /api/system/bootProgress` | Get kernel boot progress |

---

## Notebooks

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan-cli notebook list` | `POST /api/notebook/lsNotebooks` | List all notebooks |
| `siyuan-cli notebook get --id <id>` | `POST /api/notebook/getNotebookConf` | Get notebook configuration |
| `siyuan-cli notebook create --name <name>` | `POST /api/notebook/createNotebook` | Create new notebook |
| `siyuan-cli notebook rename --id <id> --name <name>` | `POST /api/notebook/renameNotebook` | Rename notebook |
| `siyuan-cli notebook remove --id <id> --yes` | `POST /api/notebook/removeNotebook` | Remove notebook |
| `siyuan-cli notebook open --id <id>` | `POST /api/notebook/openNotebook` | Open closed notebook |
| `siyuan-cli notebook close --id <id>` | `POST /api/notebook/closeNotebook` | Close open notebook |

---

## Documents

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan-cli doc list <notebook>` | `POST /api/filetree/listDocTree` | List document tree |
| `siyuan-cli doc get --path <path>` | `POST /api/export/exportMdContent` | Get document as Markdown |
| `siyuan-cli doc create --path <path> --content <md>` | `POST /api/filetree/createDocWithMd` | Create document from Markdown |
| `siyuan-cli doc rename --old-path <path> --new-path <path>` | `POST /api/filetree/renameDoc` | Rename document |
| `siyuan-cli doc move --old-path <path> --new-path <path>` | `POST /api/filetree/moveDocs` | Move document |
| `siyuan-cli doc remove --path <path> --yes` | `POST /api/filetree/removeDoc` | Remove document |

**Note:** Rename/remove by ID uses `/api/filetree/renameDocByID` and `/api/filetree/removeDocByID`

---

## Blocks

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan-cli block get --id <id>` | `POST /api/block/getBlockKramdown` | Get block kramdown |
| `siyuan-cli block children --id <id>` | `POST /api/block/getChildBlocks` | Get child blocks |
| `siyuan-cli block update --id <id> --content <md>` | `POST /api/block/updateBlock` | Update block content |
| `siyuan-cli block insert --parent-id <id> --content <md>` | `POST /api/block/insertBlock` | Insert new block |
| `siyuan-cli block remove --id <id> --yes` | `POST /api/block/deleteBlock` | Delete block |

---

## Search

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan-cli search block <keyword>` | `POST /api/search/fullTextSearchBlock` | Full-text search blocks |
| `siyuan-cli search doc <keyword>` | `POST /api/filetree/searchDocs` | Search documents |

---

## SQL Query

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan-cli sql query <statement>` | `POST /api/query/sql` | Execute SQL query (SELECT only) |

---

## Tags

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan-cli tag list` | `POST /api/tag/getTag` | List all tags |
| `siyuan-cli tag search <keyword>` | `POST /api/search/searchTag` | Search tags |
| `siyuan-cli tag docs --label <tag>` | `POST /api/tag/getTag` + filter | Get documents by tag |

---

## Attributes

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan-cli attr list` | `POST /api/attr/getBlockAttrs` (multiple) | List all attributes |
| `siyuan-cli attr get --id <id>` | `POST /api/attr/getBlockAttrs` | Get block attributes |
| `siyuan-cli attr set --id <id> --key <k> --value <v>` | `POST /api/attr/setBlockAttrs` | Set block attribute |
| `siyuan-cli attr reset --id <id> --key <k>` | `POST /api/attr/setBlockAttrs` (empty value) | Reset attribute |

---

## Export

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan-cli export preview <id>` | `POST /api/export/exportMdContent` | Preview document |
| `siyuan-cli export markdown <id>` | `POST /api/export/exportMdContent` | Export as Markdown |
| `siyuan-cli export html <id>` | `POST /api/export/exportHTML` | Export as HTML |
| `siyuan-cli export pdf <id>` | `POST /api/export/exportPDF` | Export as PDF (server-side) |
| `siyuan-cli export docx <id>` | `POST /api/export/exportDocx` | Export as DOCX (server-side) |

---

## Templates

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan-cli template list` | `POST /api/file/readDir` (path: `/data/templates`) | List templates |
| `siyuan-cli template get <path>` | `POST /api/file/getFile` | Get template content |
| `siyuan-cli template render <path> --id <doc-id>` | `POST /api/template/render` | Render template to document |
| `siyuan-cli template remove <path> --yes` | `POST /api/file/removeFile` | Remove template |

---

## Snapshots (Repository)

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan-cli snapshot list` | `POST /api/repo/getRepoSnapshots` | List all snapshots |
| `siyuan-cli snapshot current` | `POST /api/repo/getRepoSnapshot` | Get current snapshot |
| `siyuan-cli snapshot create --memo <memo>` | `POST /api/repo/createRepoSnapshot` | Create snapshot |
| `siyuan-cli snapshot restore <id> --yes` | `POST /api/repo/checkoutRepo` | Restore to snapshot |
| `siyuan-cli snapshot remove <id> --yes` | `POST /api/repo/removeRepoSnapshot` | Remove snapshot |

---

## Assets

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan-cli asset upload <file>` | `POST /api/asset/upload` | Upload asset file |
| `siyuan-cli asset unused` | `POST /api/asset/getUnusedAssets` | List unused assets |
| `siyuan-cli asset clean --yes` | `POST /api/asset/removeUnusedAssets` | Remove unused assets |

---

## File Operations

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan-cli file tree <path>` | `POST /api/file/readDir` | List directory contents |
| `siyuan-cli file read <path>` | `POST /api/file/getFile` | Read file content |
| `siyuan-cli file write <path> --content <text>` | `POST /api/file/putFile` | Write file content |
| `siyuan-cli file mkdir <path>` | `POST /api/file/putFile` (isDir: true) | Create directory |
| `siyuan-cli file remove <path> --yes` | `POST /api/file/removeFile` | Remove file/directory |
| `siyuan-cli file rename <old> <new>` | `POST /api/file/renameFile` | Rename/move file |

---

## Summary by API Category

| Category | Endpoints Used | Commands |
|----------|----------------|----------|
| System | 3 | `system *` |
| Notebook | 7 | `notebook *` |
| Filetree | 7 | `doc *`, `search doc` |
| Block | 5 | `block *` |
| Search | 2 | `search *`, `tag search` |
| SQL | 1 | `sql query` |
| Tag | 1 | `tag list` |
| Attribute | 2 | `attr *` |
| Export | 5 | `export *` |
| Template | 2 | `template list`, `template get`, `template remove` |
| File | 5 | `file *`, `template *`, `asset upload` |
| Asset | 3 | `asset *` |
| Repository | 4 | `snapshot *` |

---

## Authentication

All API endpoints require authentication header:
```
Authorization: Token {SIYUAN_TOKEN}
```

Base URL: `{SIYUAN_BASE_URL}` (default: `http://127.0.0.1:6806`)
