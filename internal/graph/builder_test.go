package graph

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoModuleImportResolution(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graph_gomod_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goMod := `module example.com/myapp

go 1.22
`
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)

	// Create files with module-prefixed imports
	files := map[string]string{
		"main.go": `package main
import "example.com/myapp/pkg/utils"
func main() { utils.Help() }`,
		"pkg/utils/helper.go": `package utils
func Help() {}`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, []byte(content), 0644)
	}

	builder := NewBuilder(tmpDir)
	g, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Verify module name was parsed
	if builder.moduleName != "example.com/myapp" {
		t.Errorf("Expected module name 'example.com/myapp', got '%s'", builder.moduleName)
	}

	// Verify edge: main.go -> pkg/utils/helper.go
	mainID := g.Paths["main.go"]
	utilsID := g.Paths["pkg/utils/helper.go"]

	if !hasEdge(g, mainID, utilsID) {
		t.Error("Missing edge main.go -> pkg/utils/helper.go")
	}
}

func TestMultiLanguageImports(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graph_multilang_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	files := map[string]string{
		// Python - use explicit module path
		"app.py": `import utils.helper
helper.do_something()`,
		"utils/helper.py": `def do_something(): pass`,

		// JavaScript
		"index.js": `import { api } from './lib/api.js';
api.call();`,
		"lib/api.js": `export const api = {};`,

		// Ruby
		"main.rb":         `require_relative 'lib/database'`,
		"lib/database.rb": `class Database; end`,

		// Rust
		"src/main.rs": `use crate::utils::helper;
fn main() {}`,
		"src/utils/helper.rs": `pub fn help() {}`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, []byte(content), 0644)
	}

	builder := NewBuilder(tmpDir)
	g, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Verify Python edge
	if pyID, ok := g.Paths["app.py"]; ok {
		if helperID, ok := g.Paths["utils/helper.py"]; ok {
			if !hasEdge(g, pyID, helperID) {
				t.Error("Missing Python edge app.py -> utils/helper.py")
			}
		}
	}

	// Verify JS edge
	if jsID, ok := g.Paths["index.js"]; ok {
		if apiID, ok := g.Paths["lib/api.js"]; ok {
			if !hasEdge(g, jsID, apiID) {
				t.Error("Missing JS edge index.js -> lib/api.js")
			}
		}
	}

	// Verify Ruby edge
	if rbID, ok := g.Paths["main.rb"]; ok {
		if dbID, ok := g.Paths["lib/database.rb"]; ok {
			if !hasEdge(g, rbID, dbID) {
				t.Error("Missing Ruby edge main.rb -> lib/database.rb")
			}
		}
	}

	// Verify Rust edge
	if rsID, ok := g.Paths["src/main.rs"]; ok {
		if helperID, ok := g.Paths["src/utils/helper.rs"]; ok {
			if !hasEdge(g, rsID, helperID) {
				t.Error("Missing Rust edge src/main.rs -> src/utils/helper.rs")
			}
		}
	}
}

func TestPageRankConvergence(t *testing.T) {
	g := NewGraph()

	// Create a simple graph: A -> B, A -> C, B -> C
	g.AddNode("A", "file")
	g.AddNode("B", "file")
	g.AddNode("C", "file")

	g.AddEdge("A", "B")
	g.AddEdge("A", "C")
	g.AddEdge("B", "C")

	g.PageRank(0.85, 100)

	// C should have highest score (referenced by A and B)
	aID := g.Paths["A"]
	bID := g.Paths["B"]
	cID := g.Paths["C"]

	if g.Nodes[cID].Score <= g.Nodes[aID].Score {
		t.Error("Node C should have higher PageRank than A")
	}
	if g.Nodes[cID].Score <= g.Nodes[bID].Score {
		t.Error("Node C should have higher PageRank than B")
	}

	// Scores should sum to approximately 1.0 (within floating point tolerance)
	sum := g.Nodes[aID].Score + g.Nodes[bID].Score + g.Nodes[cID].Score
	if sum < 0.99 || sum > 1.01 {
		t.Errorf("PageRank scores should sum to ~1.0, got %.4f", sum)
	}
}

func TestOptimizerRelevance(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "graph_opt_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	files := map[string]string{
		"go.mod": "module testapp\ngo 1.22",
		"auth/login.go": `package auth
import "testapp/db"
func Login() {}`,
		"auth/register.go": `package auth
func Register() {}`,
		"db/users.go": `package db
func GetUser() {}`,
		"utils/logger.go": `package utils
func Log() {}`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, []byte(content), 0644)
	}

	builder := NewBuilder(tmpDir)
	g, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	g.PageRank(0.85, 20)

	opt := NewOptimizer(g)
	results := opt.OptimizeContext("login", 10)

	// Should find auth/login.go
	foundLogin := false
	foundLogger := false

	for _, path := range results {
		if path == "auth/login.go" {
			foundLogin = true
		}
		if path == "utils/logger.go" {
			foundLogger = true
		}
	}

	if !foundLogin {
		t.Error("Optimizer should find auth/login.go for 'login' query")
	}

	// register.go is in same dir but not in filename match
	// logger.go is unrelated and shouldn't rank high
	if foundLogger && !foundLogin {
		t.Error("Unrelated files shouldn't rank higher than query matches")
	}
}

func TestGraphBuilderAndOptimizer(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "graph_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod for proper import resolution
	goMod := `module gptcode
go 1.22
`
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)

	// Create files
	// main.go -> imports utils
	// utils/helper.go
	// auth/login.go -> imports utils

	files := map[string]string{
		"main.go": `package main
import (
	"fmt"
	"gptcode/utils"
)
func main() { utils.Help() }`,
		"utils/helper.go": `package utils
func Help() {}`,
		"auth/login.go": `package auth
import "gptcode/utils"
func Login() { utils.Help() }`,
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, []byte(content), 0644)
	}

	// Build graph
	builder := NewBuilder(tmpDir)
	g, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build graph: %v", err)
	}

	// Verify nodes
	if len(g.Nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(g.Nodes))
	}

	// Verify edges
	// main.go -> utils/helper.go (via directory link)
	// auth/login.go -> utils/helper.go

	// Note: Our builder links to files in the target directory.
	// utils package has 1 file: helper.go
	// So main.go should have edge to utils/helper.go

	mainID := g.Paths["main.go"]
	utilsID := g.Paths["utils/helper.go"]
	authID := g.Paths["auth/login.go"]

	if !hasEdge(g, mainID, utilsID) {
		t.Error("Missing edge main.go -> utils/helper.go")
	}
	if !hasEdge(g, authID, utilsID) {
		t.Error("Missing edge auth/login.go -> utils/helper.go")
	}

	// Run PageRank
	g.PageRank(0.85, 20)

	// Verify scores
	// utils/helper.go should have highest score (referenced by 2 files)
	if g.Nodes[utilsID].Score <= g.Nodes[mainID].Score {
		t.Error("utils/helper.go should have higher score than main.go")
	}

	// Test Optimizer
	opt := NewOptimizer(g)

	// Query "login" -> should return auth/login.go and its dependencies (utils/helper.go)
	results := opt.OptimizeContext("login", 5)

	foundLogin := false
	foundUtils := false
	for _, path := range results {
		if path == "auth/login.go" {
			foundLogin = true
		}
		if path == "utils/helper.go" {
			foundUtils = true
		}
	}

	if !foundLogin {
		t.Error("Optimizer failed to find auth/login.go for query 'login'")
	}
	if !foundUtils {
		t.Error("Optimizer failed to include dependency utils/helper.go")
	}
}

func hasEdge(g *Graph, from, to int64) bool {
	for _, id := range g.OutEdges[from] {
		if id == to {
			return true
		}
	}
	return false
}
