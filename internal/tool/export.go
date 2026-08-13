package tool

import (
	"context"
	"fmt"
	"os"
)

func runExport(ctx context.Context, request Request) Response {
	input, err := decodeInput(request)
	if err != nil {
		return errorResponse("export", InvalidRequest, "input is not valid JSON", "send an input object", false)
	}
	client, toolError := clientOrError()
	if toolError != nil {
		return Response{Version: ProtocolVersion, Tool: "export", OK: false, Summary: toolError.Message, Warnings: []Warning{}, NextActions: []NextAction{}, Error: toolError}
	}
	if request.Operation == "preview_document" {
		id := stringInput(input, "document_id")
		// Resolve through the block read endpoint so a bad document ID is a
		// structured NOT_FOUND instead of a silent empty preview.
		if _, err := client.GetBlockKramdown(ctx, id); err != nil {
			if isNotFoundError(err) {
				return errorResponse("export", NotFound, "document was not found", "check the document ID and retry", false)
			}
			return remoteResponse("export", err)
		}
		doc, err := client.GetDocumentContent(ctx, id)
		if err != nil {
			return remoteResponse("export", err)
		}
		return success("export", "document preview", map[string]any{"id": id, "hPath": doc.HPath, "content": doc.Content})
	}
	return sideEffect("export", request, input, func() ([]Target, error) { return stateTargets(ctx, client, request, input) }, func() (any, error) {
		id, outputPath := stringInput(input, "document_id"), stringInput(input, "output_path")
		switch request.Operation {
		case "export_markdown":
			content, err := client.GetDocumentContent(ctx, id)
			if err != nil {
				return nil, err
			}
			canonical, pathErr := canonicalLocalPath(outputPath)
			if pathErr != nil {
				return nil, pathErr
			}
			if writeErr := os.WriteFile(canonical, []byte(content.Content), 0644); writeErr != nil {
				return nil, &localIOError{err: fmt.Errorf("cannot write Markdown export %q: %w", canonical, writeErr)}
			}
			return map[string]any{"document_id": id, "output_path": canonical, "bytes": len(content.Content)}, nil
		case "export_html":
			content, err := client.ExportHTML(ctx, id)
			if err != nil {
				return nil, err
			}
			canonical, pathErr := canonicalLocalPath(outputPath)
			if pathErr != nil {
				return nil, pathErr
			}
			if writeErr := os.WriteFile(canonical, []byte(content.Content), 0644); writeErr != nil {
				return nil, &localIOError{err: fmt.Errorf("cannot write HTML export %q: %w", canonical, writeErr)}
			}
			return map[string]any{"document_id": id, "output_path": canonical, "bytes": len(content.Content)}, nil
		case "export_pdf":
			result, err := client.ExportPDF(ctx, id, canonicalRemotePath(outputPath))
			if err != nil {
				return nil, err
			}
			return map[string]any{"document_id": id, "output_path": result.Path}, nil
		case "export_docx":
			result, err := client.ExportDocx(ctx, id, canonicalRemotePath(outputPath))
			if err != nil {
				return nil, err
			}
			return map[string]any{"document_id": id, "output_path": result.Path}, nil
		default:
			return nil, fmtNotFound("operation")
		}
	})
}
