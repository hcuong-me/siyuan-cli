// Package tool records the migration contract for the replacement CLI.
package tool

import "testing"

type legacyBehavior struct {
	legacy    string
	tool      string
	operation string
	endpoint  string
}

// legacyBehaviors is deliberately explicit: adding or removing a legacy leaf
// command requires an intentional migration decision.
var legacyBehaviors = []legacyBehavior{
	{"system version", "status", "version", "/api/system/version"},
	{"system time", "status", "time", "/api/system/currentTime"},
	{"system boot-progress", "status", "boot_progress", "/api/system/bootProgress"},
	{"notebook list", "context", "list_notebooks", "/api/notebook/lsNotebooks"},
	{"notebook create", "organize", "create_notebook", "/api/notebook/createNotebook"},
	{"notebook rename", "organize", "rename_notebook", "/api/notebook/renameNotebook"},
	{"notebook remove", "organize", "remove_notebook", "/api/notebook/removeNotebook"},
	{"notebook open", "organize", "open_notebook", "/api/notebook/openNotebook"},
	{"notebook close", "organize", "close_notebook", "/api/notebook/closeNotebook"},
	{"doc list", "context", "list_documents", "/api/filetree/listDocsByPath (recursive)"},
	{"doc get", "context", "read_document", "/api/filetree/getIDsByHPath + /api/export/exportMdContent"},
	{"doc create", "write", "create_document", "/api/filetree/createDocWithMd"},
	{"doc update", "write", "update_document", "/api/filetree/getIDsByHPath + /api/block/appendBlock"},
	{"doc remove", "organize", "remove_document", "/api/filetree/removeDoc"},
	{"block get", "context", "read_block", "/api/block/getBlockKramdown"},
	{"block children", "context", "list_block_children", "/api/block/getChildBlocks"},
	{"block update", "write", "update_block", "/api/block/updateBlock"},
	{"block append", "write", "append_block", "/api/block/insertBlock"},
	{"block insert-after", "write", "insert_after_block", "/api/block/insertBlock"},
	{"block delete", "organize", "delete_block", "/api/block/deleteBlock"},
	{"block move", "organize", "move_block", "/api/block/moveBlock"},
	{"search block", "context", "search_blocks", "/api/search/fullTextSearchBlock"},
	{"search doc", "context", "search_documents", "/api/filetree/searchDocs"},
	{"tag list", "context", "list_tags", "/api/tag/getTag"},
	{"tag search", "context", "search_tags", "/api/search/searchTag"},
	{"tag rename", "organize", "rename_tag", "/api/tag/renameTag"},
	{"tag remove", "organize", "remove_tag", "/api/tag/removeTag"},
	{"attr get", "context", "get_attributes", "/api/attr/getBlockAttrs"},
	{"attr set", "write", "set_attribute", "/api/attr/setBlockAttrs"},
	{"attr set-multiple", "write", "set_attributes", "/api/attr/setBlockAttrs"},
	{"attr reset", "write", "reset_attribute", "/api/attr/setBlockAttrs"},
	{"attr bookmarks", "context", "list_bookmarks", "/api/attr/getBookmarkLabels"},
	{"export preview", "export", "preview_document", "/api/export/exportMdContent"},
	{"export markdown", "export", "export_markdown", "/api/export/exportMdContent"},
	{"export html", "export", "export_html", "/api/export/exportHTML"},
	{"export pdf", "export", "export_pdf", "/api/export/exportPDF"},
	{"export docx", "export", "export_docx", "/api/export/exportDocx"},
	{"sql query", "context", "query", "/api/query/sql"},
	{"template list", "maintain", "list_templates", "/api/file/readDir"},
	{"template get", "maintain", "get_template", "/api/file/getFile"},
	{"template render", "maintain", "render_template", "/api/template/render"},
	{"template remove", "maintain", "remove_template", "/api/file/removeFile"},
	{"snapshot list", "maintain", "list_snapshots", "/api/repo/getRepoSnapshots"},
	{"snapshot create", "maintain", "create_snapshot", "/api/repo/createSnapshot"},
	{"snapshot restore", "maintain", "restore_snapshot", "/api/repo/checkoutRepo"},
	{"asset upload", "maintain", "upload_asset", "/api/asset/upload"},
	{"asset unused", "maintain", "list_unused_assets", "/api/asset/getUnusedAssets"},
	{"asset clean", "maintain", "clean_unused_assets", "/api/asset/removeUnusedAssets"},
	{"file tree", "maintain", "read_tree", "/api/file/readDir"},
	{"file read", "maintain", "read_file", "/api/file/getFile"},
	{"file write", "maintain", "write_file", "/api/file/putFile"},
	{"file mkdir", "maintain", "make_directory", "/api/file/putFile"},
	{"file remove", "maintain", "remove_file", "/api/file/removeFile"},
	{"file rename", "maintain", "rename_file", "/api/file/renameFile"},
}

func TestLegacyBehaviorCoverage(t *testing.T) {
	const expectedLeafCommands = 54
	if len(legacyBehaviors) != expectedLeafCommands {
		t.Fatalf("mapped %d legacy leaf commands; want %d", len(legacyBehaviors), expectedLeafCommands)
	}

	seen := make(map[string]bool, len(legacyBehaviors))
	groups := make(map[string]bool)
	catalog := Catalog()
	for _, behavior := range legacyBehaviors {
		if behavior.legacy == "" || behavior.tool == "" || behavior.operation == "" || behavior.endpoint == "" {
			t.Fatalf("incomplete migration mapping: %#v", behavior)
		}
		if seen[behavior.legacy] {
			t.Fatalf("duplicate legacy mapping: %s", behavior.legacy)
		}
		seen[behavior.legacy] = true
		toolSchema, ok := catalog[behavior.tool]
		if !ok {
			t.Fatalf("%s maps to unknown tool %q", behavior.legacy, behavior.tool)
		}
		if _, ok := toolSchema.Operations[behavior.operation]; !ok {
			t.Fatalf("%s maps to unknown operation %s.%s", behavior.legacy, behavior.tool, behavior.operation)
		}
		for i := 0; i < len(behavior.legacy); i++ {
			if behavior.legacy[i] == ' ' {
				groups[behavior.legacy[:i]] = true
				break
			}
		}
	}

	const expectedGroups = 13
	if len(groups) != expectedGroups {
		t.Fatalf("mapped %d command groups; want %d", len(groups), expectedGroups)
	}
}
