# Legacy command baseline

This file freezes the source-derived behavior of the CLI before the agent-oriented replacement. It maps all 57 registered leaf commands. The replacement has no legacy aliases.

All commands accepted the root JSON-output flag. Commands that read a path also accepted local content files where noted. “Multi-call” identifies selector resolution or per-item work that the replacement must preserve as a single operation result.

| Legacy command | Inputs and current behavior | Replacement | Endpoint or multi-call |
| --- | --- | --- | --- |
| `system version` | none; read | `status.version` | `/api/system/version` |
| `system time` | none; read | `status.time` | `/api/system/currentTime` |
| `system boot-progress` | none; read | `status.boot_progress` | `/api/system/bootProgress` |
| `notebook` / `list` | none; read | `context.list_notebooks` | `/api/notebook/lsNotebooks` |
| `notebook create` | name; mutation | `organize.create_notebook` | `/api/notebook/createNotebook` |
| `notebook rename` | id, name; mutation | `organize.rename_notebook` | `/api/notebook/renameNotebook` |
| `notebook` / `remove` | id, confirmation flag; destructive | `organize.remove_notebook` | `/api/notebook/removeNotebook` |
| `notebook open` | id; mutation | `organize.open_notebook` | `/api/notebook/openNotebook` |
| `notebook close` | id; mutation | `organize.close_notebook` | `/api/notebook/closeNotebook` |
| `doc list` | notebook id, `--max-depth`; read tree | `context.list_documents` | `/api/filetree/listDocsByPath` walked recursively from `/` |
| `doc` / `get` | notebook id, required path flag; read Markdown | `context.read_document` | `getIDsByHPath`, then `exportMdContent` |
| `doc create` | notebook id, required `--path`, `--content` or `--content-file`; mutation | `write.create_document` | `/api/filetree/createDocWithMd` |
| `doc update` | notebook id, required path flag, content or content-file; appends Markdown, despite command name | `write.update_document` | `getIDsByHPath`, then `/api/block/appendBlock`; replacement uses replace semantics |
| `doc` / `remove` | notebook id, required path flag, confirmation flag; destructive | `organize.remove_document` | resolve ID, then `/api/filetree/removeDocByID` |
| `block get` | block id; read | `context.read_block` | `/api/block/getBlockKramdown` |
| `block children` | block id; read | `context.list_block_children` | `/api/block/getChildBlocks` |
| `block update` | block id, data, `--data-type`; mutation | `write.update_block` | `/api/block/updateBlock` |
| `block append` | parent id, data, `--data-type`; mutation | `write.append_block` | `/api/block/insertBlock` |
| `block insert-after` | previous id, data, `--data-type`; mutation | `write.insert_after_block` | `/api/block/insertBlock` |
| `block` / `delete` | block id, confirmation flag; destructive | `organize.delete_block` | `/api/block/deleteBlock` |
| `block move` | block id, parent id, optional `--previous-id`; mutation | `organize.move_block` | `/api/block/moveBlock` |
| `search block` | keyword plus search flags; read | `context.search_blocks` | `/api/search/fullTextSearchBlock` |
| `search doc` | keyword plus search flags; read | `context.search_documents` (complete, unpaginated result set) | `/api/filetree/searchDocs` |
| `tag list` | none; read | `context.list_tags` | `/api/tag/getTag` |
| `tag search` | keyword; read | `context.search_tags` | `/api/search/searchTag` |
| `tag docs` | tag name; read | dropped: `/api/tag/getDocsByTag` was removed from SiYuan; use `context.search_tags` | — |
| `tag rename` | old name, new name; mutation | `organize.rename_tag` | `/api/tag/renameTag` |
| `tag` / `remove` | name, confirmation flag; destructive | `organize.remove_tag` | `/api/tag/removeTag` |
| `attr get` | block id; read | `context.get_attributes` | `/api/attr/getBlockAttrs` |
| `attr set` | block id, name, value; mutation | `write.set_attribute` | `/api/attr/setBlockAttrs` |
| `attr set-multiple` | block id and JSON attributes; mutation | `write.set_attributes` | `/api/attr/setBlockAttrs` |
| `attr reset` | block id, name; mutation | `write.reset_attribute` | `/api/attr/setBlockAttrs` with an empty value |
| `attr bookmarks` | none; read | `context.list_bookmarks` | `/api/attr/getBookmarkLabels` |
| `export preview` | doc id; Markdown read | `export.preview_document` | `/api/export/exportMdContent` |
| `export markdown` | doc id, optional `--output`; read or local file write | `export.export_markdown` | `/api/export/exportMdContent` |
| `export html` | doc id, optional `--output`; read or local file write | `export.export_html` | `/api/export/exportHTML` |
| `export pdf` | doc id, optional server-side output; server export | `export.export_pdf` | `/api/export/exportPDF` |
| `export docx` | doc id, optional server-side output; server export | `export.export_docx` | `/api/export/exportDocx` |
| `sql query` | SQL statement; read-only validation before query | `context.query` | `/api/query/sql` |
| `template list` | none; read | `maintain.list_templates` | `/api/file/readDir` |
| `template get` | path; read | `maintain.get_template` | `/api/file/getFile` |
| `template render` | path plus JSON context; read/render | `maintain.render_template` | `/api/template/render` |
| `template` / `remove` | path, confirmation flag; destructive | `maintain.remove_template` | `/api/file/removeFile` |
| `snapshot list` | none; read | `maintain.list_snapshots` | `/api/repo/getRepoSnapshots` (requires `page`) |
| `snapshot current` | none; read | dropped: `/api/repo/getRepoSnapshot` was removed from SiYuan | — |
| `snapshot create` | name, optional description; mutation | `maintain.create_snapshot` | `/api/repo/createSnapshot` |
| `snapshot` / `restore` | id, confirmation flag; destructive | `maintain.restore_snapshot` | `/api/repo/checkoutRepo` |
| `snapshot` / `remove` | id, confirmation flag; destructive | dropped: `/api/repo/removeRepoSnapshot` was removed from SiYuan | — |
| `asset upload` | local file, optional `--assets-dir`; local read and upload | `maintain.upload_asset` | `/api/asset/upload` |
| `asset unused` | none; read | `maintain.list_unused_assets` | `/api/asset/getUnusedAssets` |
| `asset` / `clean` | confirmation flag; destructive | `maintain.clean_unused_assets` | `/api/asset/removeUnusedAssets` |
| `file tree` | path, optional `--depth`; read | `maintain.read_tree` | `/api/file/readDir` |
| `file read` | path; read | `maintain.read_file` | `/api/file/getFile` |
| `file write` | path plus content or `--content-file`; mutation | `maintain.write_file` | `/api/file/putFile` |
| `file mkdir` | path; mutation | `maintain.make_directory` | `/api/file/putFile` |
| `file` / `remove` | path, confirmation flag; destructive | `maintain.remove_file` | `/api/file/removeFile` |
| `file rename` | old path, new path; mutation | `maintain.rename_file` | `/api/file/renameFile` |

The exported tool protocol replaces the local confirmation flag with `mode: "preview"` and `mode: "apply"`. Preview does not make a mutation or generate a PDF or DOCX. The protocol returns one JSON envelope for every replacement tool invocation.
