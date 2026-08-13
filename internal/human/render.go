package human

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"siyuan/internal/tool"
)

// Render converts one tool response into readable text.
func Render(response tool.Response, operation string) string {
	if response.Error != nil {
		return renderError(response.Error)
	}
	if isPreviewData(response.Data) {
		return renderPreview(response.Data)
	}
	if text, ok := renderKnown(response.Tool, operation, response.Data); ok {
		return text
	}
	return renderFallback(response.Data)
}

func renderError(err *tool.Error) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Error: %s\n", err.Message)
	if err.Fix != "" {
		fmt.Fprintf(&builder, "Fix: %s\n", err.Fix)
	}
	fmt.Fprintf(&builder, "Code: %s\n", err.Code)
	if len(err.Candidates) > 0 {
		builder.WriteString("Candidates:\n")
		for _, candidate := range err.Candidates {
			fmt.Fprintf(&builder, "  %s", candidate.ID)
			if candidate.Name != "" {
				fmt.Fprintf(&builder, "  %s", candidate.Name)
			}
			if candidate.Path != "" {
				fmt.Fprintf(&builder, "  %s", candidate.Path)
			}
			builder.WriteByte('\n')
		}
	}
	return builder.String()
}

func renderKnown(toolName, operation string, data any) (string, bool) {
	switch toolName + "." + operation {
	case "status.version":
		return fmt.Sprintf("Server version: %v\n", field(data, "version")), true
	case "status.time":
		// SiYuan returns epoch milliseconds, not seconds.
		millis, _ := field(data, "time").(float64)
		stamp := time.Unix(int64(millis)/1000, 0).UTC().Format("2006-01-02 15:04:05 MST")
		return fmt.Sprintf("Server time: %s\n", stamp), true
	case "status.boot_progress":
		progress, _ := field(data, "progress").(float64)
		details, _ := field(data, "details").(string)
		return fmt.Sprintf("Boot progress: %d%% %s\n", int(progress), details), true
	case "context.resolve_notebook":
		return renderNotebook(data), true
	case "context.list_notebooks":
		return renderNotebooks(data), true
	case "context.list_documents":
		return renderTree(data), true
	case "context.read_document":
		return renderReadDocument(data), true
	case "context.read_block":
		return fmt.Sprintf("Block %v\n\n%s\n", field(data, "id"), stringOr(field(data, "kramdown"), "")), true
	case "context.list_block_children":
		return renderChildren(data), true
	case "context.search_blocks":
		return renderBlockSearch(data), true
	case "context.search_documents":
		return renderDocumentSearch(data), true
	case "context.list_tags":
		return renderTagList(data), true
	case "context.search_tags":
		return renderTagSearch(data), true
	case "context.get_attributes":
		return renderAttributes(data), true
	case "context.list_bookmarks":
		return renderBookmarks(data), true
	case "context.query":
		return renderQuery(data), true
	case "export.preview_document":
		return fmt.Sprintf("Document: %v\n\n%s\n", field(data, "hPath"), stringOr(field(data, "content"), "")), true
	case "maintain.get_template", "maintain.read_file", "maintain.render_template":
		return fmt.Sprintf("Path: %v\n\n%s\n", field(data, "path"), stringOr(field(data, "content"), "")), true
	case "maintain.list_templates":
		return renderTemplateList(data), true
	case "maintain.read_tree":
		return renderFileList(data), true
	case "maintain.list_snapshots":
		return renderSnapshotList(data), true
	case "maintain.list_unused_assets":
		return renderAssetList(data), true
	}
	return "", false
}

func renderNotebook(data any) string {
	m := asMap(data)
	return fmt.Sprintf("Notebook: %v (%v)\n", m["name"], m["id"])
}

func renderNotebooks(data any) string {
	m := asMap(data)
	items, _ := m["notebooks"].([]any)
	var builder strings.Builder
	fmt.Fprintf(&builder, "Notebooks (%d):\n", len(items))
	for _, item := range items {
		notebook := asMap(item)
		status := ""
		if closed, _ := notebook["closed"].(bool); closed {
			status = " (closed)"
		}
		fmt.Fprintf(&builder, "  %-12v %v%s\n", notebook["id"], notebook["name"], status)
	}
	return builder.String()
}

func renderTree(data any) string {
	m := asMap(data)
	nodes, _ := m["tree"].([]any)
	var builder strings.Builder
	var walk func(items []any, depth int)
	walk = func(items []any, depth int) {
		for _, item := range items {
			node := asMap(item)
			fmt.Fprintf(&builder, "%s%v\n", strings.Repeat("  ", depth), node["name"])
			if children, ok := node["children"].([]any); ok {
				walk(children, depth+1)
			}
		}
	}
	walk(nodes, 0)
	return builder.String()
}

func renderReadDocument(data any) string {
	m := asMap(data)
	content := asMap(m["document"])
	return fmt.Sprintf("Document\n  ID: %v\n  Notebook: %v\n  Path: %v\n\n%s\n",
		m["id"], m["notebook"], m["path"], stringOr(content["content"], ""))
}

func renderChildren(data any) string {
	m := asMap(data)
	items, _ := m["children"].([]any)
	var builder strings.Builder
	fmt.Fprintf(&builder, "Child blocks (%d):\n", len(items))
	for _, item := range items {
		child := asMap(item)
		fmt.Fprintf(&builder, "  %v  %v\n", child["id"], child["type"])
	}
	return builder.String()
}

func renderBlockSearch(data any) string {
	m := asMap(data)
	blocks, _ := m["blocks"].([]any)
	count, _ := m["matchedBlockCount"].(float64)
	pages, _ := m["pageCount"].(float64)
	var builder strings.Builder
	fmt.Fprintf(&builder, "%d matching blocks (%d pages)\n", int(count), int(pages))
	for _, item := range blocks {
		block := asMap(item)
		fmt.Fprintf(&builder, "  %v\n    %v\n    id %v\n", block["hPath"], block["content"], block["id"])
	}
	return builder.String()
}

func renderDocumentSearch(data any) string {
	m := asMap(data)
	count, _ := m["count"].(float64)
	documents, _ := m["documents"].([]any)
	var builder strings.Builder
	fmt.Fprintf(&builder, "%d documents\n", int(count))
	for _, item := range documents {
		doc := asMap(item)
		fmt.Fprintf(&builder, "  %v  %v  (%v)\n", doc["id"], doc["hPath"], doc["box"])
	}
	return builder.String()
}

func renderTagList(data any) string {
	m := asMap(data)
	tags, _ := m["tags"].([]any)
	var builder strings.Builder
	fmt.Fprintf(&builder, "Tags (%d):\n", len(tags))
	for _, item := range tags {
		tag := asMap(item)
		fmt.Fprintf(&builder, "  %v (%v)\n", tag["label"], tag["count"])
	}
	return builder.String()
}

func renderTagSearch(data any) string {
	m := asMap(data)
	var builder strings.Builder
	fmt.Fprintf(&builder, "Tags matching %q:\n", stringOr(m["k"], ""))
	if tags, ok := m["tags"].([]any); ok {
		for _, tag := range tags {
			fmt.Fprintf(&builder, "  %v\n", tag)
		}
	}
	if blocks, ok := m["blocks"].([]any); ok && len(blocks) > 0 {
		builder.WriteString("Blocks:\n")
		for _, item := range blocks {
			block := asMap(item)
			fmt.Fprintf(&builder, "  %v  %v\n", block["hPath"], block["content"])
		}
	}
	return builder.String()
}

func renderAttributes(data any) string {
	m := asMap(data)
	if len(m) == 0 {
		return "No attributes.\n"
	}
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var builder strings.Builder
	for _, key := range keys {
		fmt.Fprintf(&builder, "  %v: %v\n", key, m[key])
	}
	return builder.String()
}

func renderBookmarks(data any) string {
	m := asMap(data)
	bookmarks, _ := m["bookmarks"].([]any)
	var builder strings.Builder
	fmt.Fprintf(&builder, "Bookmarks (%d):\n", len(bookmarks))
	for _, item := range bookmarks {
		bookmark := asMap(item)
		fmt.Fprintf(&builder, "  %v (%v)\n", bookmark["label"], bookmark["count"])
	}
	return builder.String()
}

func renderQuery(data any) string {
	m := asMap(data)
	rows, _ := m["rows"].([]any)
	var builder strings.Builder
	fmt.Fprintf(&builder, "%d rows\n", len(rows))
	for _, item := range rows {
		row := asMap(item)
		keys := make([]string, 0, len(row))
		for key := range row {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			parts = append(parts, fmt.Sprintf("%s=%v", key, row[key]))
		}
		fmt.Fprintf(&builder, "  %s\n", strings.Join(parts, " "))
	}
	return builder.String()
}

func renderFileList(data any) string {
	items := asSlice(data)
	var builder strings.Builder
	fmt.Fprintf(&builder, "%d entries\n", len(items))
	for _, item := range items {
		entry := asMap(item)
		marker := "file"
		if isDir, _ := entry["isDir"].(bool); isDir {
			marker = "dir"
		}
		fmt.Fprintf(&builder, "  [%s] %v\n", marker, entry["name"])
	}
	return builder.String()
}

func renderSnapshotList(data any) string {
	m := asMap(data)
	items, _ := m["snapshots"].([]any)
	var builder strings.Builder
	fmt.Fprintf(&builder, "Snapshots (%d):\n", len(items))
	for _, item := range items {
		snapshot := asMap(item)
		fmt.Fprintf(&builder, "  %v  %v\n", snapshot["id"], snapshot["memo"])
	}
	return builder.String()
}

func renderTemplateList(data any) string {
	m := asMap(data)
	items, _ := m["templates"].([]any)
	var builder strings.Builder
	fmt.Fprintf(&builder, "Templates (%d):\n", len(items))
	for _, item := range items {
		entry := asMap(item)
		marker := "file"
		if isDir, _ := entry["isDir"].(bool); isDir {
			marker = "dir"
		}
		fmt.Fprintf(&builder, "  [%s] %v\n", marker, entry["name"])
	}
	return builder.String()
}

func renderAssetList(data any) string {
	m := asMap(data)
	items, _ := m["assets"].([]any)
	var builder strings.Builder
	fmt.Fprintf(&builder, "Unused assets (%d):\n", len(items))
	for _, item := range items {
		asset := asMap(item)
		fmt.Fprintf(&builder, "  %v\n", asset["path"])
	}
	return builder.String()
}

func renderPreview(data any) string {
	m := asMap(data)
	preview := asMap(m["preview"])
	confirmation := asMap(m["confirmation"])
	var builder strings.Builder
	builder.WriteString("Preview\n")
	if targets, ok := preview["targets"].([]any); ok && len(targets) > 0 {
		builder.WriteString("  Targets:\n")
		for _, item := range targets {
			target := asMap(item)
			fmt.Fprintf(&builder, "    %v\n", target["id"])
			if details, ok := target["details"].(map[string]any); ok && len(details) > 0 {
				keys := make([]string, 0, len(details))
				for key := range details {
					keys = append(keys, key)
				}
				sort.Strings(keys)
				for _, key := range keys {
					fmt.Fprintf(&builder, "      %s: %v\n", key, details[key])
				}
			}
		}
	}
	if changes := preview["changes"]; changes != nil {
		builder.WriteString("  Changes:\n")
		encoded, _ := json.MarshalIndent(changes, "    ", "  ")
		builder.WriteString(string(encoded))
		builder.WriteString("\n")
	}
	if effects, ok := preview["irreversible_effects"].([]any); ok && len(effects) > 0 {
		builder.WriteString("  Irreversible effects:\n")
		for _, effect := range effects {
			fmt.Fprintf(&builder, "    %v\n", effect)
		}
	}
	if token, _ := confirmation["token"].(string); token != "" {
		fmt.Fprintf(&builder, "Confirmation token: %s\n", token)
	}
	if note, _ := confirmation["note"].(string); note != "" {
		fmt.Fprintf(&builder, "%s\n", note)
	}
	return builder.String()
}

func renderFallback(data any) string {
	if m := asMap(data); m != nil {
		keys := make([]string, 0, len(m))
		for key := range m {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		var builder strings.Builder
		for _, key := range keys {
			fmt.Fprintf(&builder, "%s: %v\n", key, formatValue(m[key]))
		}
		return builder.String()
	}
	if items := asSlice(data); items != nil {
		var builder strings.Builder
		for _, item := range items {
			fmt.Fprintf(&builder, "%v\n", formatValue(item))
		}
		return builder.String()
	}
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "OK\n"
	}
	return string(encoded) + "\n"
}

func formatValue(value any) string {
	if value == nil {
		return ""
	}
	switch value.(type) {
	case string, float64, bool:
		return fmt.Sprintf("%v", value)
	default:
		encoded, _ := json.Marshal(value)
		return string(encoded)
	}
}

func isPreviewData(data any) bool {
	m := asMap(data)
	if m == nil {
		return false
	}
	_, hasPreview := m["preview"]
	_, hasConfirmation := m["confirmation"]
	return hasPreview && hasConfirmation
}

func asMap(data any) map[string]any {
	if data == nil {
		return nil
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	var result map[string]any
	if json.Unmarshal(encoded, &result) != nil {
		return nil
	}
	return result
}

func asSlice(data any) []any {
	if data == nil {
		return nil
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	var result []any
	if json.Unmarshal(encoded, &result) != nil {
		return nil
	}
	return result
}

func field(data any, name string) any {
	m := asMap(data)
	if m == nil {
		return nil
	}
	return m[name]
}

func stringOr(value any, fallback string) string {
	if text, ok := value.(string); ok {
		return text
	}
	return fallback
}
