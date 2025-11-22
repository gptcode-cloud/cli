package graph

import (
	"bufio"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Builder handles graph construction
type Builder struct {
	graph      *Graph
	root       string
	cache      *Cache
	moduleName string
}

// NewBuilder creates a new graph builder
func NewBuilder(root string) *Builder {
	return &Builder{
		graph: NewGraph(),
		root:  root,
		cache: NewCache(),
	}
}

// Build scans the directory and builds the dependency graph
func (b *Builder) Build() (*Graph, error) {
	// Try to parse go.mod to get module name
	b.parseGoMod()

	if cached, err := b.cache.Get(b.root); err == nil {
		return cached, nil
	}

	err := filepath.Walk(b.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip hidden directories and vendor
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Process files based on extension
		ext := filepath.Ext(path)
		switch ext {
		case ".go":
			b.processGoFile(path)
		case ".py":
			b.processPythonFile(path)
		case ".js", ".ts", ".jsx", ".tsx":
			b.processJSFile(path)
		case ".rb":
			b.processRubyFile(path)
		case ".rs":
			b.processRustFile(path)
		}

		return nil
	})

	if err == nil {
	_ = b.cache.Set(b.root, b.graph)
	}

	return b.graph, err
}

func (b *Builder) processGoFile(path string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return
	}

	relPath, _ := filepath.Rel(b.root, path)
	b.graph.AddNode(relPath, "file")

	for _, imp := range f.Imports {
		// Clean import path (remove quotes)
		impPath := strings.Trim(imp.Path.Value, "\"")

		// Check if it's an internal import
		isInternal := false
		var targetDir string

		if b.moduleName != "" && strings.HasPrefix(impPath, b.moduleName+"/") {
			isInternal = true
			targetDir = strings.TrimPrefix(impPath, b.moduleName+"/")
		} else if !strings.Contains(impPath, ".") {
			// Fallback for subdirectories without go.mod or relative imports
			isInternal = true
			targetDir = impPath
		}

		if isInternal {
			fullTargetDir := filepath.Join(b.root, targetDir)

			if _, err := os.Stat(fullTargetDir); err == nil {
				// Find all .go files in that directory
				files, _ := filepath.Glob(filepath.Join(fullTargetDir, "*.go"))
				for _, targetFile := range files {
					targetRel, _ := filepath.Rel(b.root, targetFile)
					b.graph.AddEdge(relPath, targetRel)
				}
			}
		}
	}
}

func (b *Builder) parseGoMod() {
	goModPath := filepath.Join(b.root, "go.mod")
	file, err := os.Open(goModPath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			b.moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
			return
		}
	}
}

var (
	pyImportRegex   = regexp.MustCompile(`^(?:from\s+([\w\.]+)|import\s+([\w\.]+))`)
	jsImportRegex   = regexp.MustCompile(`(?:import|require)\s*.*?['"]([^'"]+)['"]`)
	rubyImportRegex = regexp.MustCompile(`^(?:require|require_relative)\s+['"]([^'"]+)['"]`)
	rustUseRegex    = regexp.MustCompile(`^use\s+([\w:]+)`)
)

func (b *Builder) processPythonFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	relPath, _ := filepath.Rel(b.root, path)
	b.graph.AddNode(relPath, "file")

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		matches := pyImportRegex.FindStringSubmatch(line)
		if len(matches) > 0 {
			// matches[1] is for 'from X', matches[2] is for 'import X'
			imp := matches[1]
			if imp == "" {
				imp = matches[2]
			}
			if imp == "" {
				continue
			}

			// Convert python dot notation to path
			impPath := strings.ReplaceAll(imp, ".", "/")

			// Try multiple resolution strategies
			// 1. Relative to current file's directory
			targetPath := filepath.Join(filepath.Dir(path), impPath+".py")
			if _, err := os.Stat(targetPath); err == nil {
				targetRel, _ := filepath.Rel(b.root, targetPath)
				b.graph.AddEdge(relPath, targetRel)
				continue
			}

			// 2. Absolute from root
			targetPath = filepath.Join(b.root, impPath+".py")
			if _, err := os.Stat(targetPath); err == nil {
				targetRel, _ := filepath.Rel(b.root, targetPath)
				b.graph.AddEdge(relPath, targetRel)
				continue
			}

			// 3. Try __init__.py in directory
			targetPath = filepath.Join(b.root, impPath, "__init__.py")
			if _, err := os.Stat(targetPath); err == nil {
				targetRel, _ := filepath.Rel(b.root, targetPath)
				b.graph.AddEdge(relPath, targetRel)
			}
		}
	}
}

func (b *Builder) processJSFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	relPath, _ := filepath.Rel(b.root, path)
	b.graph.AddNode(relPath, "file")

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		matches := jsImportRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			imp := matches[1]

			// Handle relative imports
			if strings.HasPrefix(imp, ".") {
				dir := filepath.Dir(path)
				targetPath := filepath.Join(dir, imp)

				// Try extensions
				exts := []string{".js", ".ts", ".jsx", ".tsx", "/index.js", "/index.ts"}
				for _, ext := range exts {
					check := targetPath
					if !strings.HasSuffix(targetPath, ext) && !strings.HasSuffix(targetPath, "/") {
						check = targetPath + ext
					}

					if _, err := os.Stat(check); err == nil {
						targetRel, _ := filepath.Rel(b.root, check)
						b.graph.AddEdge(relPath, targetRel)
						break
					}
				}
			}
		}
	}
}

func (b *Builder) processRubyFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	relPath, _ := filepath.Rel(b.root, path)
	b.graph.AddNode(relPath, "file")

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		matches := rubyImportRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			imp := matches[1]

			// Convert to path
			var targetPath string
			if strings.Contains(line, "require_relative") {
				// Relative import
				dir := filepath.Dir(path)
				targetPath = filepath.Join(dir, imp+".rb")
			} else {
				// Absolute from root
				targetPath = filepath.Join(b.root, "lib", imp+".rb")
			}

			if _, err := os.Stat(targetPath); err == nil {
				targetRel, _ := filepath.Rel(b.root, targetPath)
				b.graph.AddEdge(relPath, targetRel)
			}
		}
	}
}

func (b *Builder) processRustFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	relPath, _ := filepath.Rel(b.root, path)
	b.graph.AddNode(relPath, "file")

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		matches := rustUseRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			imp := matches[1]

			// Simple heuristic: if starts with "crate::" it's local
			if strings.HasPrefix(imp, "crate::") || strings.HasPrefix(imp, "super::") {
				// Convert module path to file path
				modPath := strings.TrimPrefix(imp, "crate::")
				modPath = strings.TrimPrefix(modPath, "super::")
				modPath = strings.ReplaceAll(modPath, "::", "/")

				// Try .rs file or mod.rs
				for _, suffix := range []string{".rs", "/mod.rs"} {
					targetPath := filepath.Join(b.root, "src", modPath+suffix)
					if _, err := os.Stat(targetPath); err == nil {
						targetRel, _ := filepath.Rel(b.root, targetPath)
						b.graph.AddEdge(relPath, targetRel)
						break
					}
				}
			}
		}
	}
}
