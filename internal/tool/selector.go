package tool

import (
	"context"
	"sort"

	"siyuan/internal/siyuan"
)

// DocumentResolution is the canonical identity of a human-readable document
// selector. The path remains in the response for v1 compatibility, while all
// mutations use the resolved ID.
type DocumentResolution struct {
	ID       string
	Notebook string
	Path     string
}

// resolveDocument resolves exactly one (notebook, path) pair. It never picks
// the first result when SiYuan returns duplicate IDs.
func resolveDocument(ctx context.Context, client *siyuan.Client, notebook, documentPath string) (DocumentResolution, error) {
	return resolveDocumentForRequest(ctx, client, Request{Mode: "preview"}, notebook, documentPath)
}

func resolveDocumentForRequest(ctx context.Context, client *siyuan.Client, request Request, notebook, documentPath string) (DocumentResolution, error) {
	// Resolve the notebook selector so a name binds to the same identity as
	// its box ID, and the API never receives a name in the notebook field.
	resolved, err := resolveNotebookForRequest(ctx, client, request, notebook)
	if err != nil {
		return DocumentResolution{}, err
	}
	ids, err := client.GetIDsByHPath(ctx, resolved.ID, documentPath)
	if err != nil {
		return DocumentResolution{}, err
	}
	ids = append([]string(nil), ids...)
	sort.Strings(ids)
	switch len(ids) {
	case 0:
		return DocumentResolution{}, notFoundTarget(request, "document", documentPath)
	case 1:
		return DocumentResolution{ID: ids[0], Notebook: resolved.ID, Path: documentPath}, nil
	default:
		candidates := make([]Candidate, 0, len(ids))
		for _, id := range ids {
			candidates = append(candidates, Candidate{ID: id, Path: documentPath})
		}
		return DocumentResolution{}, ambiguousTarget("document", documentPath, candidates)
	}
}
