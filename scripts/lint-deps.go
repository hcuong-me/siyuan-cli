// lint-deps verifies the repository's internal package dependency direction.
//go:build ignore

package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const modulePath = "siyuan"

type layer string

const (
	configLayer layer = "internal/config"
	siyuanLayer layer = "internal/siyuan"
	logicLayer  layer = "internal/logic"
	utilsLayer  layer = "internal/utils"
	toolLayer   layer = "internal/tool"
	humanLayer  layer = "internal/human"
	entryLayer  layer = "cmd"
)

var allowed = map[layer]map[layer]bool{
	configLayer: {configLayer: true},
	siyuanLayer: {siyuanLayer: true, configLayer: true},
	logicLayer:  {logicLayer: true, siyuanLayer: true, configLayer: true},
	utilsLayer:  {utilsLayer: true},
	toolLayer:   {toolLayer: true, logicLayer: true, siyuanLayer: true, configLayer: true},
	humanLayer:  {humanLayer: true, toolLayer: true},
	entryLayer:  {entryLayer: true, configLayer: true, siyuanLayer: true, logicLayer: true, utilsLayer: true, toolLayer: true, humanLayer: true},
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		fail(fmt.Sprintf("WHAT: cannot determine the repository root: %v\nWHY: dependency boundaries cannot be evaluated without source paths.\nHOW: run this check from the repository root.", err))
	}

	files, err := goFiles(root, true)
	if err != nil {
		fail(err.Error())
	}

	var violations []string
	for _, file := range files {
		from := classify(relativeDir(root, file))
		if from == "" {
			continue
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			violations = append(violations, fmt.Sprintf("WHAT: cannot parse %s: %v\nWHY: imports must be readable to enforce package boundaries.\nHOW: fix the Go syntax in this file, then rerun go run scripts/lint-deps.go.", relativeFile(root, file), err))
			continue
		}
		for _, spec := range parsed.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil || (path != modulePath && !strings.HasPrefix(path, modulePath+"/")) {
				continue
			}
			to := classify(strings.TrimPrefix(path, modulePath+"/"))
			if to != "" && !allowed[from][to] {
				violations = append(violations, fmt.Sprintf("WHAT: %s imports %q (%s -> %s).\nWHY: %s may depend only on %s; this import reverses the repository's dependency direction.\nHOW: move the caller to an allowed layer, pass the needed value through an interface or parameter, or relocate shared code to the owning layer.", relativeFile(root, file), path, from, to, from, allowedLayers(from)))
			}
		}
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		fail(strings.Join(violations, "\n\n"))
	}
}

func goFiles(root string, skipTests bool) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "vendor") {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || (skipTests && strings.HasSuffix(entry.Name(), "_test.go")) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("WHAT: cannot scan Go files: %v\nWHY: dependency boundaries must cover every production Go file.\nHOW: restore read access to the repository and rerun go run scripts/lint-deps.go", err)
	}
	sort.Strings(files)
	return files, nil
}

func relativeDir(root, file string) string {
	return filepath.ToSlash(filepath.Dir(strings.TrimPrefix(file, root+string(filepath.Separator))))
}
func relativeFile(root, file string) string {
	return filepath.ToSlash(strings.TrimPrefix(file, root+string(filepath.Separator)))
}

func classify(path string) layer {
	for _, candidate := range []layer{configLayer, siyuanLayer, logicLayer, utilsLayer, toolLayer, humanLayer, entryLayer} {
		if path == string(candidate) || strings.HasPrefix(path, string(candidate)+"/") {
			return candidate
		}
	}
	return ""
}

func allowedLayers(from layer) string {
	var layers []string
	for candidate := range allowed[from] {
		layers = append(layers, string(candidate))
	}
	sort.Strings(layers)
	return strings.Join(layers, ", ")
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
