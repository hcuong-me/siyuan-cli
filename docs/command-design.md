# siyuan-cli Command to API Mapping

This document maps CLI commands to their corresponding SiYuan API endpoints.

---

## System

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan system version` | `POST /api/system/version` | Get SiYuan kernel version |
| `siyuan system time` | `POST /api/system/currentTime` | Get current system time |
| `siyuan system boot-progress` | `POST /api/system/bootProgress` | Get kernel boot progress |

---

## Notebooks

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan notebook list` | `POST /api/notebook/lsNotebooks` | List all notebooks |
| `siyuan notebook get --id <id>` | `POST /api/notebook/getNotebookConf` | Get notebook configuration |
| `siyuan notebook create --name <name>` | `POST /api/notebook/createNotebook` | Create new notebook |
| `siyuan notebook rename --id <id> --name <name>` | `POST /api/notebook/renameNotebook` | Rename notebook |
| `siyuan notebook remove --id <id> --yes` | `POST /api/notebook/removeNotebook` | Remove notebook |
| `siyuan notebook open --id <id>` | `POST /api/notebook/openNotebook` | Open closed notebook |
| `siyuan notebook close --id <id>` | `POST /api/notebook/closeNotebook` | Close open notebook |

---

## Documents

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan doc list <notebook>` | `POST /api/filetree/listDocTree` | List document tree |
| `siyuan doc get --path <path>` | `POST /api/export/exportMdContent` | Get document as Markdown |
| `siyuan doc create --path <path> --content <md>` | `POST /api/filetree/createDocWithMd` | Create document from Markdown |
| `siyuan doc rename --old-path <path> --new-path <path>` | `POST /api/filetree/renameDoc` | Rename document |
| `siyuan doc move --old-path <path> --new-path <path>` | `POST /api/filetree/moveDocs` | Move document |
| `siyuan doc remove --path <path> --yes` | `POST /api/filetree/removeDoc` | Remove document |

**Note:** Rename/remove by ID uses `/api/filetree/renameDocByID` and `/api/filetree/removeDocByID`

---

## Blocks

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan block get --id <id>` | `POST /api/block/getBlockKramdown` | Get block kramdown |
| `siyuan block children --id <id>` | `POST /api/block/getChildBlocks` | Get child blocks |
| `siyuan block update --id <id> --content <md>` | `POST /api/block/updateBlock` | Update block content |
| `siyuan block insert --parent-id <id> --content <md>` | `POST /api/block/insertBlock` | Insert new block |
| `siyuan block remove --id <id> --yes` | `POST /api/block/deleteBlock` | Delete block |

---

## Search

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan search block <keyword>` | `POST /api/search/fullTextSearchBlock` | Full-text search blocks |
| `siyuan search doc <keyword>` | `POST /api/filetree/searchDocs` | Search documents |

---

## SQL Query

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan sql query <statement>` | `POST /api/query/sql` | Execute SQL query (SELECT only) |

---

## Tags

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan tag list` | `POST /api/tag/getTag` | List all tags |
| `siyuan tag search <keyword>` | `POST /api/search/searchTag` | Search tags |
| `siyuan tag docs --label <tag>` | `POST /api/tag/getTag` + filter | Get documents by tag |

---

## Attributes

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan attr list` | `POST /api/attr/getBlockAttrs` (multiple) | List all attributes |
| `siyuan attr get --id <id>` | `POST /api/attr/getBlockAttrs` | Get block attributes |
| `siyuan attr set --id <id> --key <k> --value <v>` | `POST /api/attr/setBlockAttrs` | Set block attribute |
| `siyuan attr reset --id <id> --key <k>` | `POST /api/attr/setBlockAttrs` (empty value) | Reset attribute |

---

## Export

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan export preview <id>` | `POST /api/export/exportMdContent` | Preview document |
| `siyuan export markdown <id>` | `POST /api/export/exportMdContent` | Export as Markdown |
| `siyuan export html <id>` | `POST /api/export/exportHTML` | Export as HTML |
| `siyuan export pdf <id>` | `POST /api/export/exportPDF` | Export as PDF (server-side) |
| `siyuan export docx <id>` | `POST /api/export/exportDocx` | Export as DOCX (server-side) |

---

## Templates

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan template list` | `POST /api/file/readDir` (path: `/data/templates`) | List templates |
| `siyuan template get <path>` | `POST /api/file/getFile` | Get template content |
| `siyuan template render <path> --id <doc-id>` | `POST /api/template/render` | Render template to document |
| `siyuan template remove <path> --yes` | `POST /api/file/removeFile` | Remove template |

---

## Snapshots (Repository)

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan snapshot list` | `POST /api/repo/getRepoSnapshots` | List all snapshots |
| `siyuan snapshot current` | `POST /api/repo/getRepoSnapshot` | Get current snapshot |
| `siyuan snapshot create --memo <memo>` | `POST /api/repo/createRepoSnapshot` | Create snapshot |
| `siyuan snapshot restore <id> --yes` | `POST /api/repo/checkoutRepo` | Restore to snapshot |
| `siyuan snapshot remove <id> --yes` | `POST /api/repo/removeRepoSnapshot` | Remove snapshot |

---

## Assets

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan asset upload <file>` | `POST /api/asset/upload` | Upload asset file |
| `siyuan asset unused` | `POST /api/asset/getUnusedAssets` | List unused assets |
| `siyuan asset clean --yes` | `POST /api/asset/removeUnusedAssets` | Remove unused assets |

---

## File Operations

| CLI Command | API Endpoint | Description |
|-------------|--------------|-------------|
| `siyuan file tree <path>` | `POST /api/file/readDir` | List directory contents |
| `siyuan file read <path>` | `POST /api/file/getFile` | Read file content |
| `siyuan file write <path> --content <text>` | `POST /api/file/putFile` | Write file content |
| `siyuan file mkdir <path>` | `POST /api/file/putFile` (isDir: true) | Create directory |
| `siyuan file remove <path> --yes` | `POST /api/file/removeFile` | Remove file/directory |
| `siyuan file rename <old> <new>` | `POST /api/file/renameFile` | Rename/move file |

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
