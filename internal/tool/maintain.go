package tool

import (
	"context"
	"fmt"
	"os"
)

func runMaintain(ctx context.Context, request Request) Response {
	input, err := decodeInput(request)
	if err != nil {
		return errorResponse("maintain", InvalidRequest, "input is not valid JSON", "send an input object", false)
	}
	client, toolError := clientOrError()
	if toolError != nil {
		return Response{Version: ProtocolVersion, Tool: "maintain", OK: false, Summary: toolError.Message, Warnings: []Warning{}, NextActions: []NextAction{}, Error: toolError}
	}
	switch request.Operation {
	case "list_templates":
		value, err := client.ListTemplates(ctx)
		if err != nil {
			return remoteResponse("maintain", err)
		}
		return success("maintain", "templates", map[string]any{"count": len(value), "templates": value})
	case "get_template":
		value, err := client.GetTemplate(ctx, stringInput(input, "path"))
		if err != nil {
			return remoteResponse("maintain", err)
		}
		return success("maintain", "template", map[string]string{"path": stringInput(input, "path"), "content": value})
	case "render_template":
		value, err := client.RenderTemplate(ctx, stringInput(input, "document_id"), stringInput(input, "path"))
		if err != nil {
			return remoteResponse("maintain", err)
		}
		return success("maintain", "rendered template", value)
	case "list_snapshots":
		value, err := client.ListSnapshots(ctx)
		if err != nil {
			return remoteResponse("maintain", err)
		}
		return success("maintain", "snapshots", map[string]any{"count": len(value), "snapshots": value})
	case "list_unused_assets":
		value, err := client.GetUnusedAssets(ctx)
		if err != nil {
			return remoteResponse("maintain", err)
		}
		return success("maintain", "unused assets", map[string]any{"count": len(value), "assets": value})
	case "read_tree":
		value, err := client.ReadDir(ctx, stringInput(input, "path"))
		if err != nil {
			return remoteResponse("maintain", err)
		}
		return success("maintain", "file tree", value)
	case "read_file":
		value, err := client.GetFile(ctx, stringInput(input, "path"))
		if err != nil {
			return remoteResponse("maintain", err)
		}
		return success("maintain", "file", map[string]string{"path": stringInput(input, "path"), "content": value})
	}
	return sideEffect("maintain", request, input, func() ([]Target, error) { return stateTargets(ctx, client, request, input) }, func() (any, error) {
		switch request.Operation {
		case "remove_template":
			err := client.RemoveTemplate(ctx, stringInput(input, "path"))
			return map[string]string{"path": canonicalRemotePath(stringInput(input, "path"))}, err
		case "create_snapshot":
			name := stringInput(input, "name")
			err := client.CreateSnapshot(ctx, name)
			if err != nil {
				return nil, err
			}
			// The createSnapshot endpoint returns no identifier. Derive the
			// created snapshot ID from the snapshot list when it is readable,
			// and degrade to the memo alone otherwise.
			result := map[string]string{"name": name}
			snapshots, listErr := client.ListSnapshots(ctx)
			if listErr == nil {
				for _, snapshot := range snapshots {
					if snapshot.Memo == name {
						result["snapshot_id"] = snapshot.ID
						break
					}
				}
			}
			return result, nil
		case "restore_snapshot":
			err := client.RestoreSnapshot(ctx, stringInput(input, "snapshot_id"))
			return map[string]string{"snapshot_id": stringInput(input, "snapshot_id")}, err
		case "upload_asset":
			canonical, pathErr := canonicalLocalPath(stringInput(input, "path"))
			if pathErr != nil {
				return nil, pathErr
			}
			if _, readErr := os.ReadFile(canonical); readErr != nil {
				return nil, &localIOError{err: fmt.Errorf("cannot read upload source %q: %w", canonical, readErr)}
			}
			value, err := client.UploadAsset(ctx, canonical)
			return value, err
		case "clean_unused_assets":
			before, err := client.GetUnusedAssets(ctx)
			if err != nil {
				return nil, err
			}
			if err := client.RemoveUnusedAssets(ctx); err != nil {
				return nil, err
			}
			after, err := client.GetUnusedAssets(ctx)
			if err != nil {
				return map[string]any{"removed": len(before)}, nil
			}
			return map[string]any{"removed": len(before), "remaining": len(after)}, nil
		case "write_file":
			err := client.PutFile(ctx, stringInput(input, "path"), stringInput(input, "content"), false)
			return map[string]string{"path": canonicalRemotePath(stringInput(input, "path"))}, err
		case "make_directory":
			err := client.PutFile(ctx, stringInput(input, "path"), "", true)
			return map[string]string{"path": canonicalRemotePath(stringInput(input, "path"))}, err
		case "remove_file":
			err := client.RemoveFile(ctx, stringInput(input, "path"))
			return map[string]string{"path": canonicalRemotePath(stringInput(input, "path"))}, err
		case "rename_file":
			err := client.RenameFile(ctx, stringInput(input, "old_path"), stringInput(input, "new_path"))
			return map[string]string{
				"old_path": canonicalRemotePath(stringInput(input, "old_path")),
				"new_path": canonicalRemotePath(stringInput(input, "new_path")),
			}, err
		default:
			return nil, fmtNotFound("operation")
		}
	})
}
