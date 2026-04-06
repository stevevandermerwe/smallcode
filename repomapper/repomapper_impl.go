//go:build !linux
// +build !linux

package repomapper

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

type repoMapperNative struct {
	LanguageMap map[string]*sitter.Language
	Queries     map[string]string
}

func newRepoMapperImpl() repoMapperImpl {
	rm := &repoMapperNative{
		LanguageMap: make(map[string]*sitter.Language),
		Queries:     make(map[string]string),
	}

	// Support languages
	rm.LanguageMap[".go"] = golang.GetLanguage()
	rm.LanguageMap[".py"] = python.GetLanguage()
	rm.LanguageMap[".java"] = java.GetLanguage()
	rm.LanguageMap[".ts"] = typescript.GetLanguage()
	rm.LanguageMap[".js"] = javascript.GetLanguage()

	// Hardcoded tags.scm patterns
	rm.Queries[".go"] = `
(function_declaration name: (identifier) @name) @definition.function
(method_declaration name: (field_identifier) @name) @definition.method
(type_declaration (type_spec name: (type_identifier) @name)) @definition.type
`
	rm.Queries[".py"] = `
(function_definition name: (identifier) @name) @definition.function
(class_definition name: (identifier) @name) @definition.class
`
	rm.Queries[".java"] = `
(method_declaration name: (identifier) @name) @definition.method
(class_declaration name: (identifier) @name) @definition.class
(interface_declaration name: (identifier) @name) @definition.interface
`
	rm.Queries[".ts"] = `
(function_declaration name: (identifier) @name) @definition.function
(class_declaration name: (identifier) @name) @definition.class
(method_definition name: (property_identifier) @name) @definition.method
(interface_declaration name: (identifier) @name) @definition.interface
`
	rm.Queries[".js"] = rm.Queries[".ts"]

	return rm
}

func (rm *repoMapperNative) GenerateMap(root string) (string, error) {
	files := make(map[string]*FileInfo)

	// Step 1: Discovery and symbol extraction
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		ext := filepath.Ext(path)
		lang, ok := rm.LanguageMap[ext]
		if !ok {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip files we can't read
		}

		symbols := rm.extractSymbols(lang, rm.Queries[ext], content)
		files[path] = &FileInfo{
			Path:    path,
			Symbols: symbols,
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// Step 2: Ranking (In-Degree)
	rm.calculateRanks(files, root)

	// Step 3: Formatting
	return rm.formatTree(root, files), nil
}

func (rm *repoMapperNative) extractSymbols(lang *sitter.Language, queryStr string, content []byte) []Symbol {
	if lang == nil || queryStr == "" {
		return nil
	}

	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	tree := parser.Parse(nil, content)
	if tree == nil {
		return nil
	}

	q, err := sitter.NewQuery([]byte(queryStr), lang)
	if err != nil {
		return nil
	}

	cursor := sitter.NewQueryCursor()
	cursor.Exec(q, tree.RootNode())

	var symbols []Symbol
	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			// We look for 'name' captures
			if q.CaptureNameForId(capture.Index) == "name" {
				node := capture.Node
				name := string(content[node.StartByte():node.EndByte()])

				// Get the full line as a simple "signature"
				startPoint := node.StartPoint()
				lines := strings.Split(string(content), "\n")
				signature := ""
				if int(startPoint.Row) < len(lines) {
					signature = strings.TrimSpace(lines[startPoint.Row])
				}

				symbols = append(symbols, Symbol{
					Name:      name,
					Signature: signature,
					Line:      int(startPoint.Row) + 1,
				})
			}
		}
	}

	return symbols
}

func (rm *repoMapperNative) calculateRanks(files map[string]*FileInfo, root string) {
	// Pre-read all file contents to avoid redundant disk I/O
	contents := make(map[string]string)
	for path := range files {
		content, err := os.ReadFile(path)
		if err == nil {
			contents[path] = string(content)
		}
	}

	// Simple In-Degree: How many files reference a symbol defined in this file
	for path, info := range files {
		for _, sym := range info.Symbols {
			for otherPath, content := range contents {
				if path == otherPath {
					continue
				}

				if strings.Contains(content, sym.Name) {
					info.Rank++
				}
			}
		}
	}
}

func (rm *repoMapperNative) formatTree(root string, files map[string]*FileInfo) string {
	var sb strings.Builder

	// Get all paths and sort them
	var paths []string
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	type node struct {
		name     string
		path     string
		children map[string]*node
		isFile   bool
	}

	treeRoot := &node{name: filepath.Base(root), children: make(map[string]*node)}

	for _, p := range paths {
		rel, _ := filepath.Rel(root, p)
		parts := strings.Split(rel, string(os.PathSeparator))
		curr := treeRoot
		for i, part := range parts {
			if _, ok := curr.children[part]; !ok {
				curr.children[part] = &node{
					name:     part,
					path:     p,
					children: make(map[string]*node),
					isFile:   i == len(parts)-1,
				}
			}
			curr = curr.children[part]
		}
	}

	var printNode func(n *node, indent string)
	printNode = func(n *node, indent string) {
		if n.isFile {
			info := files[n.path]
			// Ranking Threshold: arbitrarily say Rank > 0 is "high ranking" for visibility
			if info.Rank > 0 {
				sb.WriteString(fmt.Sprintf("%s%s (Rank: %d)\n", indent, n.name, info.Rank))
				for _, sym := range info.Symbols {
					sb.WriteString(fmt.Sprintf("%s  - %s\n", indent, sym.Signature))
				}
			} else {
				sb.WriteString(fmt.Sprintf("%s%s\n", indent, n.name))
			}
		} else {
			sb.WriteString(fmt.Sprintf("%s%s/\n", indent, n.name))

			var childNames []string
			for name := range n.children {
				childNames = append(childNames, name)
			}
			sort.Strings(childNames)

			for _, name := range childNames {
				printNode(n.children[name], indent+"  ")
			}
		}
	}

	var childNames []string
	for name := range treeRoot.children {
		childNames = append(childNames, name)
	}
	sort.Strings(childNames)
	for _, name := range childNames {
		printNode(treeRoot.children[name], "")
	}

	return sb.String()
}
