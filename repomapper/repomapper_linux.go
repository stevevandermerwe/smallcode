//go:build linux
// +build linux

package repomapper

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type repoMapperStub struct{}

func newRepoMapperImpl() repoMapperImpl {
	return &repoMapperStub{}
}

// GenerateMap returns a basic file tree without symbol extraction on Linux
// (symbol extraction via tree-sitter is not available for cross-compilation)
func (rm *repoMapperStub) GenerateMap(root string) (string, error) {
	var sb strings.Builder

	type node struct {
		name     string
		path     string
		children map[string]*node
		isFile   bool
	}

	treeRoot := &node{name: filepath.Base(root), children: make(map[string]*node)}

	// Walk directory and build tree structure (without symbol extraction)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		parts := strings.Split(rel, string(os.PathSeparator))
		curr := treeRoot
		for i, part := range parts {
			if _, ok := curr.children[part]; !ok {
				curr.children[part] = &node{
					name:     part,
					path:     path,
					children: make(map[string]*node),
					isFile:   i == len(parts)-1,
				}
			}
			curr = curr.children[part]
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to walk directory: %w", err)
	}

	// Format tree
	var printNode func(n *node, indent string)
	printNode = func(n *node, indent string) {
		if n.isFile {
			sb.WriteString(fmt.Sprintf("%s%s\n", indent, n.name))
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

	return sb.String(), nil
}
