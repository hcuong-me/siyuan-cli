package tool

import (
	"context"
	"strings"

	"siyuan/internal/logic"
	"siyuan/internal/siyuan"
)

func runContext(ctx context.Context, request Request) Response {
	input, err := decodeInput(request)
	if err != nil {
		return errorResponse("context", InvalidRequest, "input is not valid JSON", "send an input object", false)
	}
	client, toolError := clientOrError()
	if toolError != nil {
		return Response{Version: ProtocolVersion, Tool: "context", OK: false, Summary: toolError.Message, Warnings: []Warning{}, NextActions: []NextAction{}, Error: toolError}
	}

	switch request.Operation {
	case "search_blocks":
		result, err := client.FullTextSearchBlock(ctx, stringInput(input, "query"), integerInput(input, "page", 1), integerInput(input, "size", 20))
		if err != nil {
			return remoteResponse("context", err)
		}
		return success("context", "block search results", result)
	case "search_documents":
		result, err := client.SearchDocs(ctx, stringInput(input, "query"), stringInput(input, "notebook"), stringInput(input, "path"))
		if err != nil {
			return remoteResponse("context", err)
		}
		// Search results expose box and human path. Resolve IDs when the
		// server can provide them, but keep the complete result set if an
		// older server does not support the follow-up endpoint.
		for i := range result {
			if result[i].ID != "" || result[i].Box == "" || result[i].HPath == "" {
				continue
			}
			hPath := result[i].HPath
			if hPath != "" && !strings.HasPrefix(hPath, "/") {
				hPath = "/" + hPath
			}
			paths := []string{hPath}
			// searchDocs commonly returns an HPath prefixed with the notebook
			// name, while getIDsByHPath expects a path relative to that notebook.
			// Try the direct form first, then the relative form without guessing
			// an ID from an ambiguous response.
			if separator := strings.IndexByte(strings.TrimPrefix(hPath, "/"), '/'); separator >= 0 {
				relative := strings.TrimPrefix(strings.TrimPrefix(hPath, "/")[separator:], "/")
				if relative != "" {
					paths = append(paths, "/"+relative)
				}
			}
			for _, candidatePath := range paths {
				ids, resolveErr := client.GetIDsByHPath(ctx, result[i].Box, candidatePath)
				if resolveErr == nil && len(ids) == 1 {
					result[i].ID = ids[0]
					break
				}
			}
		}
		return success("context", "document search results", map[string]any{"count": len(result), "documents": result})
	case "resolve_notebook":
		notebook, err := resolveNotebookForRequest(ctx, client, request, stringInput(input, "selector"))
		if err != nil {
			return responseFromError("context", err)
		}
		return success("context", "resolved notebook", notebook)
	case "list_notebooks":
		result, err := client.ListNotebooks(ctx)
		if err != nil {
			return remoteResponse("context", err)
		}
		return success("context", "notebooks", map[string]any{"count": len(result.Notebooks), "notebooks": result.Notebooks})
	case "list_documents":
		notebook, err := resolveNotebookForRequest(ctx, client, request, stringInput(input, "notebook"))
		if err != nil {
			return responseFromError("context", err)
		}
		result, err := client.ListDocTree(ctx, notebook.ID, integerInput(input, "max_depth", 0))
		if err != nil {
			return remoteResponse("context", err)
		}
		return success("context", "document tree", map[string]any{
			"notebook": notebook,
			"count":    countTree(result.Tree),
			"tree":     result.Tree,
		})
	case "read_document":
		resolution, err := resolveDocument(ctx, client, stringInput(input, "notebook"), stringInput(input, "path"))
		if err != nil {
			return responseFromError("context", err)
		}
		result, err := client.GetDocumentContent(ctx, resolution.ID)
		if err != nil {
			return remoteResponse("context", err)
		}
		return success("context", "document content", map[string]any{"id": resolution.ID, "notebook": resolution.Notebook, "path": resolution.Path, "document": result})
	case "read_block":
		result, err := client.GetBlockKramdown(ctx, stringInput(input, "block_id"))
		if err != nil {
			return remoteResponse("context", err)
		}
		return success("context", "block content", map[string]any{"id": stringInput(input, "block_id"), "kramdown": result})
	case "list_block_children":
		result, err := client.GetChildBlocks(ctx, stringInput(input, "block_id"))
		if err != nil {
			return remoteResponse("context", err)
		}
		return success("context", "block children", map[string]any{"count": len(result), "children": result})
	case "list_tags":
		result, err := client.GetTags(ctx)
		if err != nil {
			return remoteResponse("context", err)
		}
		return success("context", "tags", map[string]any{"count": len(result), "tags": result})
	case "search_tags":
		result, err := client.SearchTags(ctx, stringInput(input, "query"))
		if err != nil {
			return remoteResponse("context", err)
		}
		return success("context", "tag search results", result)
	case "get_attributes":
		result, err := client.GetBlockAttrs(ctx, stringInput(input, "block_id"))
		if err != nil {
			return remoteResponse("context", err)
		}
		return success("context", "block attributes", result)
	case "list_bookmarks":
		result, err := client.GetBookmarkLabels(ctx)
		if err != nil {
			return remoteResponse("context", err)
		}
		return success("context", "bookmarks", map[string]any{"count": len(result), "bookmarks": result})
	case "query":
		logic, err := logic.NewSQLLogic()
		if err != nil {
			return remoteResponse("context", err)
		}
		result, err := logic.Query(ctx, stringInput(input, "statement"))
		if err != nil {
			return errorResponse("context", InvalidRequest, err.Error(), "send one SELECT statement without comments or mutations", false)
		}
		return success("context", "query results", map[string]any{"count": len(result), "rows": result})
	default:
		return errorResponse("context", InvalidRequest, "unknown operation", "run tools to inspect context operations", false)
	}
}

// countTree counts every node in a document tree, including nested children.
func countTree(nodes []siyuan.TreeNode) int {
	count := 0
	for _, node := range nodes {
		count += 1 + countTree(node.Children)
	}
	return count
}
