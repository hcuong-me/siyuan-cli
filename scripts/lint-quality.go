// lint-quality enforces inexpensive source-quality rules that Go itself cannot express.
//go:build ignore

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxLines = 1000

func main() {
	root, err := os.Getwd()
	if err != nil {
		fail(fmt.Sprintf("WHAT: cannot determine the repository root: %v\nWHY: source-quality checks need stable file paths.\nHOW: run this check from the repository root.", err))
	}

	files, err := goFiles(root)
	if err != nil {
		fail(err.Error())
	}

	var violations []string
	for _, file := range files {
		violations = append(violations, checkLineCount(root, file)...)
		violations = append(violations, checkRawLogCalls(root, file)...)
		violations = append(violations, checkExportedDocComments(root, file)...)
	}
	violations = append(violations, checkSkillReferenceDrift(root)...)
	if len(violations) > 0 {
		sort.Strings(violations)
		fail(strings.Join(violations, "\n\n"))
	}
}

func goFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "vendor") {
			return filepath.SkipDir
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("WHAT: cannot scan Go files: %v\nWHY: quality rules must cover every tracked source file.\nHOW: restore read access to the repository and rerun go run scripts/lint-quality.go", err)
	}
	sort.Strings(files)
	return files, nil
}

func checkLineCount(root, file string) []string {
	contents, err := os.ReadFile(file)
	if err != nil {
		return []string{fmt.Sprintf("WHAT: cannot read %s: %v\nWHY: file size must be checked before the source can be maintained safely.\nHOW: restore read access or remove the unreadable file, then rerun go run scripts/lint-quality.go.", relativeFile(root, file), err)}
	}
	lines := strings.Count(string(contents), "\n")
	if len(contents) > 0 && !strings.HasSuffix(string(contents), "\n") {
		lines++
	}
	if lines <= maxLines {
		return nil
	}
	return []string{fmt.Sprintf("WHAT: %s has %d lines (limit: %d).\nWHY: oversized files are difficult to review and isolate in this CLI.\nHOW: extract a cohesive type or operation into a focused package or file.", relativeFile(root, file), lines, maxLines)}
}

func checkRawLogCalls(root, file string) []string {
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		return []string{fmt.Sprintf("WHAT: cannot parse %s: %v\nWHY: raw log calls cannot be identified in invalid Go source.\nHOW: fix the Go syntax in this file, then rerun go run scripts/lint-quality.go.", relativeFile(root, file), err)}
	}

	aliases, dotImported := logAliases(parsed)
	var violations []string
	ast.Inspect(parsed, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		method, receiver := "", ""
		if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
			identifier, isIdentifier := selector.X.(*ast.Ident)
			if isIdentifier && aliases[identifier.Name] && isRawLogMethod(selector.Sel.Name) {
				method, receiver = selector.Sel.Name, identifier.Name
			}
		} else if identifier, ok := call.Fun.(*ast.Ident); ok && dotImported && isRawLogMethod(identifier.Name) {
			method, receiver = identifier.Name, "log"
		}
		if method != "" {
			violations = append(violations, fmt.Sprintf("WHAT: %s:%d calls %s.%s.\nWHY: direct standard-library log output bypasses the CLI's command output conventions.\nHOW: return an error to the command layer or use internal/utils/output for user-facing output.", relativeFile(root, file), fset.Position(call.Pos()).Line, receiver, method))
		}
		return true
	})
	return violations
}

func checkExportedDocComments(root, file string) []string {
	if strings.HasSuffix(file, "_test.go") {
		return nil
	}

	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		return []string{fmt.Sprintf("WHAT: cannot parse %s: %v\nWHY: exported declarations need accurate Go doc comments.\nHOW: fix the Go syntax in this file, then rerun go run scripts/lint-quality.go.", relativeFile(root, file), err)}
	}

	var violations []string
	for _, declaration := range parsed.Decls {
		switch decl := declaration.(type) {
		case *ast.FuncDecl:
			if decl.Name.IsExported() && !hasDocComment(decl.Doc, decl.Name.Name) {
				violations = append(violations, missingDocComment(root, file, fset.Position(decl.Pos()).Line, decl.Name.Name))
			}
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch item := spec.(type) {
				case *ast.TypeSpec:
					if item.Name.IsExported() && !hasDocComment(item.Doc, item.Name.Name) && !hasDocComment(decl.Doc, item.Name.Name) {
						violations = append(violations, missingDocComment(root, file, fset.Position(item.Pos()).Line, item.Name.Name))
					}
				case *ast.ValueSpec:
					for _, name := range item.Names {
						if name.IsExported() && !hasDocComment(item.Doc, name.Name) && !hasDocComment(decl.Doc, name.Name) {
							violations = append(violations, missingDocComment(root, file, fset.Position(name.Pos()).Line, name.Name))
						}
					}
				}
			}
		}
	}
	return violations
}

// checkSkillReferenceDrift fails when a bundled skill reference has diverged
// from the docs file it mirrors. The skill carries copies so it works outside
// the repository; the copies must stay identical to their sources.
func checkSkillReferenceDrift(root string) []string {
	refs := []struct {
		reference string
		docs      string
	}{
		{"skills/siyuan-cli/references/agent-tool-contract.md", "docs/agent-tool-contract.md"},
		{"skills/siyuan-cli/references/human-interface.md", "docs/human-interface.md"},
	}
	var violations []string
	for _, pair := range refs {
		docsText, err := os.ReadFile(filepath.Join(root, pair.docs))
		if err != nil {
			continue // no canonical source to compare against
		}
		refPath := filepath.Join(root, pair.reference)
		refBytes, err := os.ReadFile(refPath)
		if err != nil {
			violations = append(violations, fmt.Sprintf("WHAT: %s is missing (source: %s).\nWHY: the skill bundles a copy of the protocol docs so it works outside the repository.\nHOW: copy %s to %s and rerun go run scripts/lint-quality.go.", pair.reference, pair.docs, pair.docs, pair.reference))
			continue
		}
		if string(docsText) == string(refBytes) {
			continue
		}
		violations = append(violations, fmt.Sprintf("WHAT: %s has drifted from %s.\nWHY: the skill must stay in sync with the protocol documentation it bundles.\nHOW: copy %s over %s and rerun go run scripts/lint-quality.go.%s", pair.reference, pair.docs, pair.docs, pair.reference, firstDivergentLines(string(docsText), string(refBytes), 3)))
	}
	return violations
}

// firstDivergentLines summarizes the first lines where two texts differ.
func firstDivergentLines(docsText, refText string, limit int) string {
	docsLines := strings.Split(docsText, "\n")
	refLines := strings.Split(refText, "\n")
	if len(docsLines) != len(refLines) {
		return fmt.Sprintf("\n  (line counts differ: docs %d, reference %d)", len(docsLines), len(refLines))
	}
	var detail []string
	for i := 0; i < len(docsLines) && len(detail) < limit; i++ {
		if docsLines[i] != refLines[i] {
			detail = append(detail, fmt.Sprintf("\n  docs line %d: %q\n  ref  line %d: %q", i+1, docsLines[i], i+1, refLines[i]))
		}
	}
	return strings.Join(detail, "")
}

func hasDocComment(group *ast.CommentGroup, name string) bool {
	if group == nil {
		return false
	}
	text := strings.TrimSpace(group.Text())
	return text == name || strings.HasPrefix(text, name+" ") || strings.HasPrefix(text, name+".")
}

func missingDocComment(root, file string, line int, name string) string {
	return fmt.Sprintf("WHAT: %s:%d exports %s without a Go doc comment that starts with its name.\nWHY: exported declarations are package contracts; callers need their purpose without reading implementation.\nHOW: add a comment immediately above the declaration, for example: // %s ...", relativeFile(root, file), line, name, name)
}

func logAliases(file *ast.File) (map[string]bool, bool) {
	aliases := map[string]bool{}
	dotImported := false
	for _, spec := range file.Imports {
		if strings.Trim(spec.Path.Value, "\"") != "log" {
			continue
		}
		name := "log"
		if spec.Name != nil {
			name = spec.Name.Name
		}
		if name == "." {
			dotImported = true
		} else if name != "_" {
			aliases[name] = true
		}
	}
	return aliases, dotImported
}

func isRawLogMethod(name string) bool {
	switch name {
	case "Print", "Printf", "Println", "Fatal", "Fatalf", "Fatalln", "Panic", "Panicf", "Panicln":
		return true
	default:
		return false
	}
}

func relativeFile(root, file string) string {
	return filepath.ToSlash(strings.TrimPrefix(file, root+string(filepath.Separator)))
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
