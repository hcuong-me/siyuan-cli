package tool

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"siyuan/internal/siyuan"
)

// ialPattern matches one kramdown inline attribute list, e.g. {: id="a" updated="b"}.
var ialPattern = regexp.MustCompile(`\{: [^}]*\}`)

// ialPairPattern matches one key="value" pair inside an inline attribute list.
var ialPairPattern = regexp.MustCompile(`\w+="[^"]*"`)

// canonicalIALs makes kramdown stable for fingerprinting. SiYuan serializes
// inline attribute lists in map order, so unchanged blocks can return
// attributes in different orders between requests.
func canonicalIALs(kramdown string) string {
	return ialPattern.ReplaceAllStringFunc(kramdown, func(match string) string {
		pairs := ialPairPattern.FindAllString(match, -1)
		if len(pairs) == 0 {
			return match
		}
		sort.Strings(pairs)
		return "{: " + strings.Join(pairs, " ") + "}"
	})
}

// sideEffect is the common preview/apply gate. The resolver is called for
// both modes, so apply always re-reads the state that was bound to preview.
func sideEffect(toolName string, request Request, input map[string]any, resolveTargets func() ([]Target, error), apply func() (any, error)) Response {
	if request.Mode == "apply" && request.ConfirmationToken == "" {
		_, toolError := Preflight(toolName, request, nil)
		return responseWithError(toolName, toolError)
	}

	targets, err := resolveTargets()
	if err != nil {
		return resolveErrorResponse(toolName, request, err)
	}
	if len(targets) == 0 {
		return responseWithError(toolName, &Error{
			Code:      PreconditionUnavailable,
			Message:   "the operation has no concrete state precondition",
			Fix:       "retry after the target state can be inspected; no confirmation token was issued",
			Retryable: true,
		})
	}

	confirmation, toolError := Preflight(toolName, request, targets)
	if toolError != nil {
		return responseWithError(toolName, toolError)
	}
	if confirmation != nil {
		return success(toolName, "preview ready", map[string]any{
			"preview": Preview{
				Targets:             targets,
				Changes:             input,
				IrreversibleEffects: irreversibleEffects(request, input, targets),
			},
			"confirmation": confirmation,
		})
	}

	data, err := apply()
	if err != nil {
		return responseFromError(toolName, err)
	}
	return success(toolName, "change applied", data)
}

// responseWithError keeps all errors in the one tool envelope.
func responseWithError(toolName string, toolError *Error) Response {
	if toolError == nil {
		toolError = &Error{Code: RemoteError, Message: "tool failed without an error", Fix: "retry the request", Retryable: true}
	}
	toolError = normalizeError(toolError)
	return Response{
		Version:     ProtocolVersion,
		Tool:        toolName,
		OK:          false,
		Summary:     toolError.Message,
		Warnings:    []Warning{},
		NextActions: defaultNextActions(toolName, "", toolError.Code),
		Error:       toolError,
	}
}

// targetResolutionError is returned by precondition resolvers. A resolver
// never silently substitutes a selector or request hash for state.
type targetResolutionError struct {
	code         string
	message      string
	fix          string
	retryable    bool
	candidates   []Candidate
	staleOnApply bool
}

// Error returns the resolution failure message.
func (e *targetResolutionError) Error() string { return e.message }

func (e *targetResolutionError) toolError() *Error {
	return &Error{
		Code:       e.code,
		Message:    e.message,
		Fix:        e.fix,
		Retryable:  e.retryable,
		Candidates: e.candidates,
	}
}

type localIOError struct{ err error }

// Error returns the local filesystem failure message.
func (e *localIOError) Error() string { return e.err.Error() }

// Unwrap exposes the underlying local filesystem failure.
func (e *localIOError) Unwrap() error { return e.err }

func localError(err error, fix string) *Error {
	return &Error{Code: LocalIOError, Message: err.Error(), Fix: fix, Retryable: true}
}

func responseFromError(toolName string, err error) Response {
	var toolErr *Error
	if errors.As(err, &toolErr) {
		return responseWithError(toolName, toolErr)
	}
	var targetErr *targetResolutionError
	if errors.As(err, &targetErr) {
		return responseWithError(toolName, targetErr.toolError())
	}
	var localErr *localIOError
	if errors.As(err, &localErr) {
		return responseWithError(toolName, localError(localErr, "check the local path and permissions, then retry"))
	}
	return remoteResponse(toolName, err)
}

func resolveErrorResponse(toolName string, request Request, err error) Response {
	var targetErr *targetResolutionError
	if errors.As(err, &targetErr) {
		if request.Mode == "apply" && targetErr.staleOnApply && targetErr.code != ConfirmationStale {
			targetErr.code = ConfirmationStale
			targetErr.message = "the confirmed target disappeared or could not be resolved"
			targetErr.fix = "run preview again and use the new token"
			targetErr.retryable = true
		}
		return responseWithError(toolName, targetErr.toolError())
	}
	var localErr *localIOError
	if errors.As(err, &localErr) {
		return responseWithError(toolName, localError(localErr, "check the local path and permissions, then retry"))
	}
	// Resolver failures are preflight failures. Keep them fail-closed and
	// distinguish them from a remote error returned by the mutation closure.
	safety := safetyFailure(request.Operation, err.Error())
	return responseWithError(toolName, safety.(*targetResolutionError).toolError())
}

func notFoundTarget(request Request, kind, selector string) error {
	if request.Mode == "apply" {
		return &targetResolutionError{
			code:         ConfirmationStale,
			message:      fmt.Sprintf("confirmed %s %q no longer exists", kind, selector),
			fix:          "run preview again and use the new token",
			retryable:    true,
			staleOnApply: true,
		}
	}
	return &targetResolutionError{
		code:      NotFound,
		message:   fmt.Sprintf("%s %q was not found", kind, selector),
		fix:       "check the selector or path and retry",
		retryable: false,
	}
}

func ambiguousTarget(kind, selector string, candidates []Candidate) error {
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].ID == candidates[j].ID {
			return candidates[i].Path < candidates[j].Path
		}
		return candidates[i].ID < candidates[j].ID
	})
	return &targetResolutionError{
		code:       AmbiguousSelector,
		message:    fmt.Sprintf("more than one %s matches %q", kind, selector),
		fix:        fmt.Sprintf("select one %s ID from the candidates", kind),
		retryable:  false,
		candidates: candidates,
	}
}

func safetyFailure(operation, detail string) error {
	return &targetResolutionError{
		code:      PreconditionUnavailable,
		message:   fmt.Sprintf("cannot safely preview %s: %s", operation, detail),
		fix:       "retry after the target state can be inspected; no confirmation token was issued",
		retryable: true,
	}
}

// stateTargets resolves all catalogued side effects to concrete, re-readable
// targets. The operation switch is deliberate: adding an operation without a
// resolver must fail closed instead of falling back to request text.
func stateTargets(ctx context.Context, client *siyuan.Client, request Request, input map[string]any) ([]Target, error) {
	switch request.Operation {
	case "create_document":
		return createDocumentTargets(ctx, client, request, input)
	case "update_document", "remove_document":
		resolution, err := resolveDocumentForRequest(ctx, client, request, stringInput(input, "notebook"), stringInput(input, "path"))
		if err != nil {
			return nil, err
		}
		content, err := client.GetDocumentContent(ctx, resolution.ID)
		if err != nil {
			if isNotFoundError(err) {
				return nil, notFoundTarget(request, "document", resolution.ID)
			}
			return nil, err
		}
		return []Target{fingerprintTargetWithDetails("document", resolution.ID, map[string]any{
			"id": resolution.ID, "notebook": resolution.Notebook, "path": resolution.Path, "content": content,
		}, map[string]string{"notebook": resolution.Notebook, "path": resolution.Path})}, nil
	case "update_block", "append_block", "insert_after_block", "delete_block", "move_block", "set_attribute", "set_attributes", "reset_attribute":
		return blockOperationTargets(ctx, client, request, input)
	case "create_notebook":
		return notebookListingTarget(ctx, client, request, stringInput(input, "name"))
	case "rename_notebook", "open_notebook", "close_notebook", "remove_notebook":
		notebook, err := resolveNotebookForRequest(ctx, client, request, stringInput(input, "notebook"))
		if err != nil {
			return nil, err
		}
		return []Target{fingerprintTargetWithDetails("notebook", notebook.ID, notebook, map[string]string{"name": notebook.Name})}, nil
	case "rename_tag", "remove_tag":
		return tagTarget(ctx, client, request, stringInput(input, "tag"))
	case "export_markdown", "export_html", "export_pdf", "export_docx":
		return exportTargets(ctx, client, request, input)
	case "remove_template":
		return remoteFileTargets(ctx, client, request, stringInput(input, "path"), "template")
	case "create_snapshot":
		return snapshotListingTarget(ctx, client)
	case "restore_snapshot":
		return snapshotTarget(ctx, client, request, stringInput(input, "snapshot_id"))
	case "upload_asset":
		return uploadTargets(ctx, client, request, stringInput(input, "path"))
	case "clean_unused_assets":
		return unusedAssetTarget(ctx, client)
	case "write_file", "make_directory", "remove_file":
		return remoteFileTargets(ctx, client, request, stringInput(input, "path"), "file")
	case "rename_file":
		oldTargets, err := remoteFileTargetsForRename(ctx, client, request, stringInput(input, "old_path"), true)
		if err != nil {
			return nil, err
		}
		newTargets, err := remoteFileTargetsForRename(ctx, client, request, stringInput(input, "new_path"), false)
		if err != nil {
			return nil, err
		}
		targets := append(oldTargets, newTargets...)
		sortTargets(targets)
		return targets, nil
	default:
		return nil, safetyFailure(request.Operation, "no operation-owned resolver is available")
	}
}

func createDocumentTargets(ctx context.Context, client *siyuan.Client, request Request, input map[string]any) ([]Target, error) {
	// Resolve the notebook selector so a name binds to the same target ID as
	// its box ID, and the API never receives a name in the notebook field.
	notebook, err := resolveNotebookForRequest(ctx, client, request, stringInput(input, "notebook"))
	if err != nil {
		return nil, err
	}
	docPath := canonicalRemotePath(stringInput(input, "path"))
	ids, err := client.GetIDsByHPath(ctx, notebook.ID, docPath)
	if err != nil {
		return nil, err
	}
	sort.Strings(ids)
	if len(ids) > 0 {
		if request.Mode == "apply" {
			return nil, &targetResolutionError{
				code:      ConfirmationStale,
				message:   "the confirmed document absence no longer holds",
				fix:       "run preview again and use the new token",
				retryable: true,
			}
		}
		return nil, &targetResolutionError{code: Conflict, message: "a document already exists at the requested path", fix: "choose a new document path", retryable: false}
	}
	return []Target{fingerprintTargetWithDetails("document-absence", notebook.ID+":"+docPath, map[string]any{
		"notebook": notebook.ID, "path": docPath, "ids": ids,
	}, map[string]string{"notebook": notebook.ID, "path": docPath})}, nil
}

func blockOperationTargets(ctx context.Context, client *siyuan.Client, request Request, input map[string]any) ([]Target, error) {
	ids := make([]string, 0, 3)
	for _, field := range []string{"block_id", "parent_id", "previous_id"} {
		if id := stringInput(input, field); id != "" && !containsString(ids, id) {
			ids = append(ids, id)
		}
	}
	targets := make([]Target, 0, len(ids))
	for _, id := range ids {
		content, err := client.GetBlockKramdown(ctx, id)
		if err != nil {
			if isNotFoundError(err) {
				return nil, notFoundTarget(request, "block", id)
			}
			return nil, err
		}
		targets = append(targets, fingerprintTargetWithDetails("block", id, canonicalIALs(content), map[string]string{"id": id}))
	}
	if len(targets) == 0 {
		return nil, safetyFailure(request.Operation, "the block target is empty")
	}
	sortTargets(targets)
	return targets, nil
}

func resolveNotebookForRequest(ctx context.Context, client *siyuan.Client, request Request, selector string) (siyuan.Notebook, error) {
	result, err := client.ListNotebooks(ctx)
	if err != nil {
		return siyuan.Notebook{}, err
	}
	matches := make([]siyuan.Notebook, 0, 1)
	for _, notebook := range result.Notebooks {
		if notebook.ID == selector || notebook.Name == selector {
			matches = append(matches, notebook)
		}
	}
	if len(matches) == 0 {
		return siyuan.Notebook{}, notFoundTarget(request, "notebook", selector)
	}
	if len(matches) > 1 {
		candidates := make([]Candidate, 0, len(matches))
		for _, notebook := range matches {
			candidates = append(candidates, Candidate{ID: notebook.ID, Name: notebook.Name})
		}
		return siyuan.Notebook{}, ambiguousTarget("notebook", selector, candidates)
	}
	return matches[0], nil
}

func notebookListingTarget(ctx context.Context, client *siyuan.Client, request Request, requestedName string) ([]Target, error) {
	result, err := client.ListNotebooks(ctx)
	if err != nil {
		return nil, err
	}
	notebooks := append([]siyuan.Notebook(nil), result.Notebooks...)
	for _, notebook := range notebooks {
		if requestedName != "" && notebook.Name == requestedName {
			if request.Mode == "apply" {
				return nil, &targetResolutionError{
					code:      ConfirmationStale,
					message:   "the confirmed notebook name is no longer available",
					fix:       "run preview again and use the new token",
					retryable: true,
				}
			}
			return nil, &targetResolutionError{code: Conflict, message: "a notebook already has the requested name", fix: "choose a different notebook name", retryable: false}
		}
	}
	sort.Slice(notebooks, func(i, j int) bool { return notebooks[i].ID < notebooks[j].ID })
	return []Target{fingerprintTarget("notebook-list", "all", notebooks)}, nil
}

// resolveTagForRequest resolves exactly one tag by its label. Tag identity is
// the label itself, but the resolution still distinguishes a missing tag from
// an ambiguous one and keeps preview and apply on the same identity path.
func resolveTagForRequest(ctx context.Context, client *siyuan.Client, request Request, selector string) (siyuan.Tag, error) {
	tags, err := client.GetTags(ctx)
	if err != nil {
		return siyuan.Tag{}, err
	}
	matches := make([]siyuan.Tag, 0, 1)
	for _, tag := range tags {
		if tag.Label == selector {
			matches = append(matches, tag)
		}
	}
	if len(matches) == 0 {
		return siyuan.Tag{}, notFoundTarget(request, "tag", selector)
	}
	if len(matches) > 1 {
		candidates := make([]Candidate, 0, len(matches))
		for _, tag := range matches {
			candidates = append(candidates, Candidate{ID: tag.Label})
		}
		return siyuan.Tag{}, ambiguousTarget("tag", selector, candidates)
	}
	return matches[0], nil
}

func tagTarget(ctx context.Context, client *siyuan.Client, request Request, selector string) ([]Target, error) {
	tag, err := resolveTagForRequest(ctx, client, request, selector)
	if err != nil {
		return nil, err
	}
	return []Target{fingerprintTarget("tag", tag.Label, tag)}, nil
}

func snapshotListingTarget(ctx context.Context, client *siyuan.Client) ([]Target, error) {
	snapshots, err := client.ListSnapshots(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(snapshots, func(i, j int) bool { return snapshots[i].ID < snapshots[j].ID })
	return []Target{fingerprintTarget("snapshot-list", "all", snapshots)}, nil
}

func snapshotTarget(ctx context.Context, client *siyuan.Client, request Request, id string) ([]Target, error) {
	snapshots, err := client.ListSnapshots(ctx)
	if err != nil {
		return nil, err
	}
	for _, snapshot := range snapshots {
		if snapshot.ID == id {
			return []Target{fingerprintTarget("snapshot", id, snapshot)}, nil
		}
	}
	return nil, notFoundTarget(request, "snapshot", id)
}

func unusedAssetTarget(ctx context.Context, client *siyuan.Client) ([]Target, error) {
	assets, err := client.GetUnusedAssets(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].Path < assets[j].Path })
	return []Target{fingerprintTarget("unused-assets", "set", assets)}, nil
}

func uploadTargets(ctx context.Context, client *siyuan.Client, request Request, source string) ([]Target, error) {
	canonical, err := canonicalLocalPath(source)
	if err != nil {
		return nil, err
	}
	state, err := localFileState(canonical, true)
	if err != nil {
		if request.Mode == "apply" && (errors.Is(err, os.ErrNotExist) || strings.Contains(strings.ToLower(err.Error()), "is a directory")) {
			return nil, &targetResolutionError{
				code:      ConfirmationStale,
				message:   fmt.Sprintf("confirmed local upload source %q changed or disappeared", canonical),
				fix:       "restore the source and run preview again",
				retryable: true,
			}
		}
		return nil, err
	}
	// Upload destinations are server-owned. Fingerprint the complete assets
	// directory listing so a new or removed destination cannot reuse a token.
	assets, err := client.ReadDir(ctx, "/data/assets")
	if err != nil {
		return nil, err
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].Name < assets[j].Name })
	return []Target{
		fingerprintTargetWithDetails("local-file", canonical, state, map[string]string{"path": canonical, "owner": "local"}),
		fingerprintTarget("asset-set", "all", assets),
	}, nil
}

func exportTargets(ctx context.Context, client *siyuan.Client, request Request, input map[string]any) ([]Target, error) {
	id := stringInput(input, "document_id")
	// Resolve export targets through the block read endpoint. This inspects
	// document state without invoking any export generator during preview.
	content, err := client.GetBlockKramdown(ctx, id)
	if err != nil {
		if isNotFoundError(err) {
			return nil, notFoundTarget(request, "document", id)
		}
		return nil, err
	}
	targets := []Target{fingerprintTargetWithDetails("document", id, canonicalIALs(content), map[string]string{"id": id})}
	destination := stringInput(input, "output_path")
	switch request.Operation {
	case "export_markdown", "export_html":
		canonical, localErr := canonicalLocalPath(destination)
		if localErr != nil {
			return nil, localErr
		}
		state, localErr := localDestinationState(canonical)
		if localErr != nil {
			if request.Mode == "apply" && errors.Is(localErr, os.ErrNotExist) {
				return nil, &targetResolutionError{
					code:      ConfirmationStale,
					message:   fmt.Sprintf("confirmed local export destination %q changed or disappeared", canonical),
					fix:       "restore the destination parent and run preview again",
					retryable: true,
				}
			}
			// A preview cannot bind a token to state it cannot inspect.
			return nil, safetyFailure(request.Operation, "the local export destination cannot be inspected: "+localErr.Error())
		}
		targets = append(targets, fingerprintTargetWithDetails("local-destination", canonical, state, map[string]string{"path": canonical, "owner": "local"}))
	case "export_pdf", "export_docx":
		remote, remoteErr := remoteFileTargets(ctx, client, request, destination, "export")
		if remoteErr != nil {
			return nil, remoteErr
		}
		targets = append(targets, remote...)
	}
	sortTargets(targets)
	return targets, nil
}

type remoteFileState struct {
	Path       string            `json:"path"`
	Exists     bool              `json:"exists"`
	IsDir      bool              `json:"is_dir"`
	Content    string            `json:"content,omitempty"`
	Parent     string            `json:"parent"`
	ParentList []siyuan.FileInfo `json:"parent_list,omitempty"`
}

func remoteFileTargets(ctx context.Context, client *siyuan.Client, request Request, rawPath, owner string) ([]Target, error) {
	requireExisting := request.Operation == "remove_file" || request.Operation == "remove_template"
	return remoteFileTargetsWithRequirement(ctx, client, request, rawPath, owner, requireExisting)
}

func remoteFileTargetsForRename(ctx context.Context, client *siyuan.Client, request Request, rawPath string, requireExisting bool) ([]Target, error) {
	return remoteFileTargetsWithRequirement(ctx, client, request, rawPath, "file", requireExisting)
}

func remoteFileTargetsWithRequirement(ctx context.Context, client *siyuan.Client, request Request, rawPath, owner string, requireExisting bool) ([]Target, error) {
	canonical := canonicalRemotePath(rawPath)
	if canonical == "." || canonical == "/" && rawPath == "" {
		return nil, safetyFailure(request.Operation, "the remote path is empty")
	}
	state, err := remoteFileStateForPath(ctx, client, canonical)
	if err != nil {
		if isNotFoundError(err) {
			return nil, notFoundTarget(request, "remote file", canonical)
		}
		return nil, err
	}
	if requireExisting && !state.Exists {
		return nil, notFoundTarget(request, "remote file", canonical)
	}
	if state.IsDir && (request.Operation == "write_file" || strings.HasPrefix(request.Operation, "export_")) {
		return nil, safetyFailure(request.Operation, "the destination is a directory")
	}
	return []Target{fingerprintTargetWithDetails("remote-destination", owner+":"+canonical, state, map[string]string{"path": canonical, "owner": "remote"})}, nil
}

func remoteFileStateForPath(ctx context.Context, client *siyuan.Client, canonical string) (remoteFileState, error) {
	content, err := client.GetFile(ctx, canonical)
	if err == nil {
		return remoteFileState{Path: canonical, Exists: true, Content: content, Parent: path.Dir(canonical)}, nil
	}
	if !isNotFoundError(err) {
		return remoteFileState{}, err
	}
	parent := path.Dir(canonical)
	entries, listErr := client.ReadDir(ctx, parent)
	if listErr != nil {
		return remoteFileState{}, listErr
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	base := path.Base(canonical)
	for _, entry := range entries {
		if entry.Name == base {
			return remoteFileState{Path: canonical, Exists: true, IsDir: entry.IsDir, Parent: parent, ParentList: entries}, nil
		}
	}
	return remoteFileState{Path: canonical, Exists: false, Parent: parent, ParentList: entries}, nil
}

func canonicalRemotePath(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	clean := path.Clean(raw)
	if !strings.HasPrefix(clean, "/") {
		clean = "/" + clean
	}
	return clean
}

func canonicalLocalPath(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", &localIOError{err: fmt.Errorf("local path is required")}
	}
	absolute, err := filepath.Abs(raw)
	if err != nil {
		return "", &localIOError{err: fmt.Errorf("cannot canonicalize local path %q: %w", raw, err)}
	}
	return filepath.Clean(absolute), nil
}

func localFileState(canonical string, requireFile bool) (map[string]any, error) {
	info, err := os.Stat(canonical)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &localIOError{err: fmt.Errorf("local source %q does not exist", canonical)}
		}
		return nil, &localIOError{err: fmt.Errorf("cannot stat local source %q: %w", canonical, err)}
	}
	if requireFile && info.IsDir() {
		return nil, &localIOError{err: fmt.Errorf("local source %q is a directory", canonical)}
	}
	state := map[string]any{"path": canonical, "mode": info.Mode().String(), "size": info.Size(), "mod_time": info.ModTime().UnixNano(), "is_dir": info.IsDir()}
	if !info.IsDir() {
		content, readErr := os.ReadFile(canonical)
		if readErr != nil {
			return nil, &localIOError{err: fmt.Errorf("cannot read local source %q: %w", canonical, readErr)}
		}
		state["content_sha256"] = hashBytes(content)
	}
	return state, nil
}

func localDestinationState(canonical string) (map[string]any, error) {
	info, err := os.Stat(canonical)
	if err == nil {
		if info.IsDir() {
			return nil, &localIOError{err: fmt.Errorf("local export destination %q is a directory", canonical)}
		}
		state, stateErr := localFileState(canonical, true)
		if stateErr != nil {
			return nil, stateErr
		}
		state["owner"] = "local"
		return state, nil
	}
	if !os.IsNotExist(err) {
		return nil, &localIOError{err: fmt.Errorf("cannot inspect local export destination %q: %w", canonical, err)}
	}
	parent := filepath.Dir(canonical)
	parentInfo, parentErr := os.Stat(parent)
	if parentErr != nil {
		return nil, &localIOError{err: fmt.Errorf("cannot inspect local export parent %q: %w", parent, parentErr)}
	}
	if !parentInfo.IsDir() {
		return nil, &localIOError{err: fmt.Errorf("local export parent %q is not a directory", parent)}
	}
	entries, readErr := os.ReadDir(parent)
	if readErr != nil {
		return nil, &localIOError{err: fmt.Errorf("cannot read local export parent %q: %w", parent, readErr)}
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return map[string]any{"path": canonical, "exists": false, "parent": parent, "parent_entries": names}, nil
}

func fingerprintTarget(kind, id string, state any) Target {
	return fingerprintTargetWithDetails(kind, id, state, nil)
}

func fingerprintTargetWithDetails(kind, id string, state any, details map[string]string) Target {
	encoded, _ := json.Marshal(state)
	sum := sha256.Sum256(encoded)
	return Target{ID: kind + ":" + id, Fingerprint: hex.EncodeToString(sum[:]), Details: details}
}

func hashBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

func sortTargets(targets []Target) {
	sort.Slice(targets, func(i, j int) bool {
		if targets[i].ID == targets[j].ID {
			return targets[i].Fingerprint < targets[j].Fingerprint
		}
		return targets[i].ID < targets[j].ID
	})
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "not found") || strings.Contains(message, "does not exist") || strings.Contains(message, "no such")
}

func irreversibleEffects(request Request, input map[string]any, targets []Target) []string {
	effects := make([]string, 0, len(targets))
	for _, target := range targets {
		id := target.ID
		kind, value := id, ""
		if index := strings.IndexByte(id, ':'); index >= 0 {
			kind, value = id[:index], id[index+1:]
		}
		effect := "change:" + id
		switch request.Operation {
		case "create_document":
			effect = "create_document:" + value
		case "update_document":
			effect = "update_document:" + value
		case "remove_document":
			effect = "delete_document:" + value
		case "update_block", "set_attribute", "set_attributes", "reset_attribute":
			effect = request.Operation + ":" + value
		case "append_block", "insert_after_block":
			effect = request.Operation + ":" + value
		case "delete_block":
			effect = "delete_block:" + value
		case "move_block":
			effect = "move_block:" + value
		case "create_notebook":
			effect = "create_notebook:" + stringInput(input, "name")
		case "remove_notebook":
			effect = "delete_notebook:" + value
		case "rename_notebook", "open_notebook", "close_notebook":
			effect = request.Operation + ":" + value
		case "rename_tag", "remove_tag":
			effect = request.Operation + ":" + value
		case "export_markdown", "export_html":
			if kind == "local-destination" {
				effect = "overwrite_local_file:" + value
			} else {
				effect = request.Operation + ":" + value
			}
		case "export_pdf", "export_docx":
			effect = request.Operation + ":" + value
		case "remove_template":
			effect = "delete_template:" + value
		case "create_snapshot":
			effect = "create_snapshot:" + stringInput(input, "name")
		case "restore_snapshot":
			effect = "restore_snapshot:" + value
		case "upload_asset":
			effect = "upload_asset:" + value
		case "clean_unused_assets":
			effect = "remove_unused_assets:" + target.Fingerprint
		case "write_file":
			effect = "overwrite_remote_file:" + value
		case "make_directory":
			effect = "create_remote_directory:" + value
		case "remove_file":
			effect = "delete_remote_file:" + value
		case "rename_file":
			effect = "rename_remote_file:" + value
		}
		effects = append(effects, effect)
	}
	sort.Strings(effects)
	return effects
}
