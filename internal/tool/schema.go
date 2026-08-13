package tool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
)

// FieldSchema describes one field allowed in an operation input object.
type FieldSchema struct {
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// OperationSchema is the single source for discovery and request validation.
type OperationSchema struct {
	Description   string                 `json:"description"`
	SideEffect    bool                   `json:"side_effect"`
	Input         map[string]FieldSchema `json:"input"`
	Example       map[string]any         `json:"example"`
	Risk          string                 `json:"risk_boundary"`
	Applicability string                 `json:"applicability,omitempty"`
	NextActions   []string               `json:"next_actions,omitempty"`
}

// Schema contains every operation exposed by one top-level tool.
type Schema struct {
	Goal       string                     `json:"goal"`
	Operations map[string]OperationSchema `json:"operations"`
}

func fields(required ...string) map[string]FieldSchema {
	result := make(map[string]FieldSchema, len(required))
	for _, name := range required {
		result[name] = FieldSchema{Type: "string", Required: true, Description: fieldDescription(name)}
	}
	return result
}

func operation(description string, sideEffect bool, required []string, risk string) OperationSchema {
	return OperationSchema{
		Description: description,
		SideEffect:  sideEffect,
		Input:       fields(required...),
		Example:     map[string]any{"version": ProtocolVersion, "operation": "", "input": map[string]any{}},
		Risk:        risk,
	}
}

func addOptional(schema OperationSchema, names ...string) OperationSchema {
	for _, name := range names {
		schema.Input[name] = FieldSchema{Type: "string", Required: false, Description: fieldDescription(name)}
	}
	return schema
}

// typed overrides the declared JSON type of the named fields.
func typed(schema OperationSchema, typ string, names ...string) OperationSchema {
	for _, name := range names {
		field := schema.Input[name]
		field.Type = typ
		schema.Input[name] = field
	}
	return schema
}

// Catalog is the complete version-one operation catalog.
func Catalog() map[string]Schema {
	read := "Runs immediately and does not change SiYuan."
	preview := "Use preview first. Apply requires the returned confirmation token."
	catalog := map[string]Schema{
		"status": {Goal: "Read server state.", Operations: map[string]OperationSchema{
			"version":       operation("Read the server version.", false, nil, read),
			"time":          operation("Read the server time.", false, nil, read),
			"boot_progress": operation("Read boot progress.", false, nil, read),
		}},
		"context": {Goal: "Find and read decision-relevant note data.", Operations: map[string]OperationSchema{
			"search_blocks":       typed(addOptional(operation("Search blocks.", false, []string{"query"}, read), "page", "size"), "integer", "page", "size"),
			"search_documents":    addOptional(operation("Search the complete document result set without pagination.", false, []string{"query"}, read), "notebook", "path"),
			"resolve_notebook":    operation("Resolve a notebook selector.", false, []string{"selector"}, read),
			"list_notebooks":      operation("List notebooks.", false, nil, read),
			"list_documents":      typed(addOptional(operation("List a notebook document tree.", false, []string{"notebook"}, read), "max_depth"), "integer", "max_depth"),
			"read_document":       operation("Read one document.", false, []string{"notebook", "path"}, read),
			"read_block":          operation("Read one block.", false, []string{"block_id"}, read),
			"list_block_children": operation("List child blocks.", false, []string{"block_id"}, read),
			"list_tags":           operation("List tags.", false, nil, read),
			"search_tags":         operation("Search tags.", false, []string{"query"}, read),
			"get_attributes":      operation("Read block attributes.", false, []string{"block_id"}, read),
			"list_bookmarks":      operation("List bookmarks.", false, nil, read),
			"query":               operation("Run a read-only SQL query.", false, []string{"statement"}, read),
		}},
		"write": {Goal: "Change document, block, and attribute content.", Operations: map[string]OperationSchema{
			"create_document":    operation("Create a Markdown document.", true, []string{"notebook", "path", "content"}, preview),
			"update_document":    operation("Replace document content.", true, []string{"notebook", "path", "content"}, preview),
			"update_block":       operation("Replace block content.", true, []string{"block_id", "content"}, preview),
			"append_block":       operation("Append a block.", true, []string{"parent_id", "content"}, preview),
			"insert_after_block": operation("Insert a block after another block.", true, []string{"previous_id", "content"}, preview),
			"set_attribute":      operation("Set one block attribute.", true, []string{"block_id", "key", "value"}, preview),
			"set_attributes":     typed(operation("Set several block attributes.", true, []string{"block_id", "attributes"}, preview), "object", "attributes"),
			"reset_attribute":    operation("Remove one block attribute.", true, []string{"block_id", "key"}, preview),
		}},
		"organize": {Goal: "Change notebook, document, tag, and block structure.", Operations: map[string]OperationSchema{
			"create_notebook": operation("Create a notebook.", true, []string{"name"}, preview),
			"rename_notebook": operation("Rename a notebook.", true, []string{"notebook", "name"}, preview),
			"open_notebook":   operation("Open a notebook.", true, []string{"notebook"}, preview),
			"close_notebook":  operation("Close a notebook.", true, []string{"notebook"}, preview),
			"remove_notebook": operation("Remove a notebook.", true, []string{"notebook"}, preview),
			"move_block":      addOptional(operation("Move a block.", true, []string{"block_id", "parent_id"}, preview), "previous_id"),
			"delete_block":    operation("Delete a block.", true, []string{"block_id"}, preview),
			"remove_document": operation("Remove a document.", true, []string{"notebook", "path"}, preview),
			"rename_tag":      operation("Rename a tag.", true, []string{"tag", "name"}, preview),
			"remove_tag":      operation("Remove a tag.", true, []string{"tag"}, preview),
		}},
		"export": {Goal: "Preview or create document exports.", Operations: map[string]OperationSchema{
			"preview_document": operation("Preview a document as Markdown.", false, []string{"document_id"}, read),
			"export_markdown":  operation("Export Markdown.", true, []string{"document_id", "output_path"}, preview),
			"export_html":      operation("Export HTML.", true, []string{"document_id", "output_path"}, preview),
			"export_pdf":       operation("Export PDF.", true, []string{"document_id", "output_path"}, preview),
			"export_docx":      operation("Export DOCX.", true, []string{"document_id", "output_path"}, preview),
		}},
		"maintain": {Goal: "Manage templates, assets, snapshots, and raw files.", Operations: map[string]OperationSchema{
			"list_templates":      operation("List templates.", false, nil, read),
			"get_template":        operation("Read a template.", false, []string{"path"}, read),
			"render_template":     operation("Render a template.", false, []string{"document_id", "path"}, read),
			"remove_template":     operation("Remove a template.", true, []string{"path"}, preview),
			"list_snapshots":      operation("List snapshots.", false, nil, read),
			"create_snapshot":     operation("Create a snapshot.", true, []string{"name"}, preview),
			"restore_snapshot":    operation("Restore a snapshot.", true, []string{"snapshot_id"}, preview),
			"upload_asset":        operation("Upload an asset.", true, []string{"path"}, preview),
			"list_unused_assets":  operation("List unused assets.", false, nil, read),
			"clean_unused_assets": operation("Remove unused assets.", true, nil, preview),
			"read_tree":           operation("Read a file tree.", false, []string{"path"}, read),
			"read_file":           operation("Read a file.", false, []string{"path"}, read),
			"write_file":          operation("Write a file.", true, []string{"path", "content"}, preview),
			"make_directory":      operation("Create a directory.", true, []string{"path"}, preview),
			"remove_file":         operation("Remove a file.", true, []string{"path"}, preview),
			"rename_file":         operation("Rename a file.", true, []string{"old_path", "new_path"}, preview),
		}},
	}
	for toolName, toolSchema := range catalog {
		for operationName, operationSchema := range toolSchema.Operations {
			operationSchema.Risk = operationRisk(operationName, operationSchema.SideEffect)
			operationSchema.Example["operation"] = operationName
			operationSchema.Example["input"] = exampleInput(operationSchema.Input)
			if operationSchema.SideEffect {
				operationSchema.Example["mode"] = "preview"
			}
			operationSchema.Applicability = applicabilityFor(operationName, operationSchema.SideEffect)
			if operationSchema.SideEffect {
				operationSchema.NextActions = []string{"Inspect preview.targets and preview.irreversible_effects, then apply with its confirmation token."}
			} else {
				operationSchema.NextActions = []string{"Use returned IDs and canonical paths for a follow-up operation."}
			}
			toolSchema.Operations[operationName] = operationSchema
		}
		catalog[toolName] = toolSchema
	}
	return catalog
}

// applicabilityFor states when an operation should and should not be used.
// Operations without a specific boundary inherit the generic read or
// side-effect contract.
func applicabilityFor(operationName string, sideEffect bool) string {
	switch operationName {
	case "search_documents":
		return "Use to find documents by text, optionally narrowed by notebook or path. Returns the complete result set without pagination. Not applicable: finding blocks inside a document (use search_blocks)."
	case "search_blocks":
		return "Use to find blocks by text. Results are paged by page/size; the server limits each page. Not applicable: listing an entire notebook (use list_documents)."
	case "query":
		return "Use for read-only SQL over the SiYuan schema. Not applicable: writes, DDL, or multi-statement scripts; each query sends exactly one SELECT statement."
	case "list_documents":
		return "Use to discover a notebook's document structure before reading or writing. notebook accepts an ID or exact name. max_depth bounds local traversal; zero returns the full tree."
	case "read_document":
		return "Use to read one document's Markdown by (notebook, path). notebook accepts an ID or exact name; path is human-readable."
	case "read_block":
		return "Use to read one block's kramdown by stable block_id. Not applicable: resolving a path to an ID (use read_document)."
	case "create_document":
		return "Use to create one Markdown document at a path inside a notebook. Returns CONFLICT when a document already exists at the path; preview confirms the absence first."
	case "update_document":
		return "Use to replace one document's content entirely. Not applicable: partial or structural edits (use update_block or organize)."
	case "remove_document":
		return "Destructive: permanently removes one document. Requires preview with a resolved target and a confirmation token."
	case "delete_block":
		return "Destructive: permanently removes one block and its children. Requires preview with a resolved target and a confirmation token."
	case "move_block":
		return "Use to re-parent one block. Optional previous_id sets the new sibling position."
	case "create_notebook":
		return "Use to create a notebook with a unique name. Returns CONFLICT when the name is already taken."
	case "remove_notebook":
		return "Destructive: removes a notebook and its documents. notebook accepts an ID or exact name; preview resolves the target."
	case "rename_tag", "remove_tag":
		return "Applies to the exact tag label across all documents. Not applicable: partial or per-block tag edits."
	case "clean_unused_assets":
		return "Destructive: removes every asset that SiYuan reports as unused. Preview fingerprints the asset set; the set must be unchanged at apply."
	case "create_snapshot":
		return "Use to snapshot the repository before a risky change. Returns the created snapshot_id when the server reports it."
	case "restore_snapshot":
		return "Destructive: replaces the current repository state with one snapshot. Requires preview and a confirmation token."
	case "export_markdown", "export_html":
		return "Writes the export to a local file after preview. output_path is local and must be writable; the parent directory must exist."
	case "export_pdf", "export_docx":
		return "Writes the export to a server-side path after preview. output_path is owned by the remote file API."
	case "upload_asset":
		return "Use to upload one local file into the server assets directory. The local source is fingerprinted at preview."
	case "write_file":
		return "Overwrites or creates one remote file under the server file API. Not applicable: document content (use write)."
	case "make_directory":
		return "Creates one remote directory under the server file API."
	case "remove_file", "remove_template":
		return "Destructive: permanently removes one remote file. Requires preview with a resolved target and a confirmation token."
	case "rename_file":
		return "Renames one remote file. The destination may be absent; preview resolves both paths."
	case "get_template":
		return "Use to read one template's raw content. Not applicable: rendering with document context (use render_template)."
	}
	if sideEffect {
		return "Use when the requested change is intentional and the resolved target is known."
	}
	return "Use for read-only inspection; it does not change SiYuan."
}

func operationRisk(operationName string, sideEffect bool) string {
	if !sideEffect {
		if operationName == "search_documents" {
			return "Read-only and unpaginated; larger result sets are returned with a count."
		}
		return "Read-only; no SiYuan or local mutation is performed."
	}
	switch operationName {
	case "remove_notebook", "remove_document", "delete_block", "remove_tag", "remove_template", "remove_file", "clean_unused_assets", "restore_snapshot":
		return "Destructive side effect; inspect resolved targets and irreversible effects before apply."
	case "export_markdown", "export_html", "export_pdf", "export_docx":
		return "Writes an export to the owner-specific local or server destination after preview."
	case "upload_asset", "write_file", "make_directory", "rename_file":
		return "Changes a file or asset destination; preview fingerprints source and destination state."
	default:
		return "State-changing operation; preview is required and apply needs its confirmation token."
	}
}

func fieldDescription(name string) string {
	switch name {
	case "notebook":
		return "Notebook ID or exact notebook name."
	case "path":
		return "Human-readable or server-owned path, depending on the operation."
	case "output_path":
		return "Destination path; Markdown and HTML are local, PDF and DOCX are server-side."
	case "document_id":
		return "Stable document block ID."
	case "block_id", "parent_id", "previous_id":
		return "Stable block ID used to resolve the target."
	case "content":
		return "Markdown or file content to write."
	case "query":
		return "Text to search for."
	case "selector":
		return "Exact ID or name selector; ambiguity is returned as candidates."
	case "max_depth":
		return "Local document-tree depth bound; zero means unlimited."
	case "attributes":
		return "JSON object whose values are string attributes."
	case "snapshot_id":
		return "Stable repository snapshot ID."
	case "old_path":
		return "Existing source path owned by the remote file API."
	case "new_path":
		return "Destination path owned by the remote file API."
	default:
		return "Value for " + name + "."
	}
}

func exampleInput(input map[string]FieldSchema) map[string]any {
	example := make(map[string]any, len(input))
	for name, field := range input {
		example[name] = exampleValue(name, field.Type)
	}
	return example
}

func exampleValue(name, typ string) any {
	if typ == "integer" {
		return 1
	}
	if typ == "object" {
		return map[string]string{"custom": "value"}
	}
	switch name {
	case "content":
		return "# Title\n\nContent"
	case "query", "statement":
		if name == "statement" {
			return "SELECT * FROM blocks LIMIT 10"
		}
		return "roadmap"
	case "notebook", "selector":
		return "20210817205410-2kvfpfn"
	case "block_id", "parent_id", "previous_id", "document_id":
		return "20210808180117-czj9bvb"
	case "path", "output_path":
		return "/Projects/Roadmap"
	case "old_path":
		return "/data/storage/old.md"
	case "new_path":
		return "/data/storage/new.md"
	case "tag":
		return "important"
	case "snapshot_id":
		return "20260801-120000"
	case "name":
		return "Example name"
	case "key":
		return "custom-key"
	case "value":
		return "custom-value"
	case "max_depth":
		return 2
	default:
		return "example-" + name
	}
}

// ValidateRequest validates a request against the catalog.
func ValidateRequest(tool string, request Request) error {
	if request.Version != ProtocolVersion {
		return fmt.Errorf("version must be %q", ProtocolVersion)
	}
	schema, ok := Catalog()[tool]
	if !ok {
		return fmt.Errorf("unknown tool %q", tool)
	}
	operation, ok := schema.Operations[request.Operation]
	if !ok {
		return fmt.Errorf("unknown operation %q for tool %q", request.Operation, tool)
	}
	if len(bytes.TrimSpace(request.Input)) == 0 {
		return fmt.Errorf("input is required")
	}

	var input map[string]json.RawMessage
	if err := json.Unmarshal(request.Input, &input); err != nil || input == nil {
		return fmt.Errorf("input must be a JSON object")
	}
	for name := range input {
		if _, ok := operation.Input[name]; !ok {
			return fmt.Errorf("input.%s is not allowed", name)
		}
	}
	for name, field := range operation.Input {
		raw, present := input[name]
		if !present {
			if field.Required {
				return fmt.Errorf("input.%s is required", name)
			}
			continue
		}
		if err := validateField(name, field, raw); err != nil {
			return err
		}
	}
	if operation.SideEffect && request.Mode != "preview" && request.Mode != "apply" {
		return fmt.Errorf("mode must be preview or apply for %s", request.Operation)
	}
	return nil
}

// validateField validates one present input value against its catalog schema.
// Validation is deliberately performed on RawMessage values so numbers do not
// pass through float64 and lose precision before their range is checked.
func validateField(name string, field FieldSchema, raw json.RawMessage) error {
	path := "input." + name
	value := bytes.TrimSpace(raw)
	if len(value) == 0 {
		return fmt.Errorf("%s is empty", path)
	}
	if bytes.Equal(value, []byte("null")) {
		return fmt.Errorf("%s must not be null", path)
	}

	switch field.Type {
	case "string":
		var decoded string
		if err := json.Unmarshal(value, &decoded); err != nil {
			return fmt.Errorf("%s must be a JSON string", path)
		}
		if field.Required && decoded == "" {
			return fmt.Errorf("%s must not be empty", path)
		}
		return nil
	case "integer":
		if _, err := parseInteger(value); err != nil {
			return fmt.Errorf("%s must be a finite integer: %v", path, err)
		}
		return nil
	case "object":
		var decoded map[string]json.RawMessage
		if err := json.Unmarshal(value, &decoded); err != nil || decoded == nil {
			return fmt.Errorf("%s must be a JSON object", path)
		}
		// set_attributes is the one declared object whose endpoint contract is
		// map[string]string. Validate its members here so the handler cannot
		// silently discard a malformed value while unmarshalling.
		if name == "attributes" {
			for key, member := range decoded {
				var stringValue string
				if bytes.Equal(bytes.TrimSpace(member), []byte("null")) || json.Unmarshal(member, &stringValue) != nil {
					return fmt.Errorf("%s.%s must be a JSON string", path, key)
				}
			}
		}
		return nil
	default:
		return fmt.Errorf("%s uses unsupported JSON type %q", path, field.Type)
	}
}

// parseInteger parses a JSON number without converting it through float64.
// It accepts integral spellings such as 1.0 and 1e2, but rejects fractions,
// non-finite values, and values outside the platform int range.
func parseInteger(raw []byte) (int, error) {
	value := bytes.TrimSpace(raw)
	if len(value) == 0 || !json.Valid(value) || (value[0] != '-' && (value[0] < '0' || value[0] > '9')) {
		return 0, fmt.Errorf("value is not a JSON number")
	}
	rational, ok := new(big.Rat).SetString(string(value))
	if !ok {
		return 0, fmt.Errorf("value is not a finite JSON number")
	}
	if !rational.IsInt() {
		return 0, fmt.Errorf("value is fractional")
	}

	var minInt64, maxInt64 int64 = -1 << 63, 1<<63 - 1
	if strconv.IntSize == 32 {
		minInt64, maxInt64 = -1<<31, 1<<31-1
	}
	minimum := big.NewInt(minInt64)
	maximum := big.NewInt(maxInt64)
	integer := rational.Num()
	if integer.Cmp(minimum) < 0 || integer.Cmp(maximum) > 0 {
		return 0, fmt.Errorf("value is outside the int range")
	}
	return int(integer.Int64()), nil
}
