package graph

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type CacheEntry struct {
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
	Graph     *Graph    `json:"graph"`
}

type Cache struct {
	cacheDir string
}

func NewCache() *Cache {
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".gptcode", "cache")
	_ = os.MkdirAll(cacheDir, 0755)
	return &Cache{cacheDir: cacheDir}
}

func (c *Cache) cacheKey(root string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(root)))
}

func (c *Cache) cachePath(root string) string {
	return filepath.Join(c.cacheDir, fmt.Sprintf("graph_%s.json", c.cacheKey(root)))
}

func (c *Cache) computeHash(root string) (string, error) {
	h := md5.New()

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "vendor" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".go" || ext == ".py" || ext == ".js" || ext == ".ts" || ext == ".jsx" || ext == ".tsx" {
			relPath, _ := filepath.Rel(root, path)
			h.Write([]byte(relPath))
			fmt.Fprintf(h, "%d", info.ModTime().Unix())
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (c *Cache) Get(root string) (*Graph, error) {
	cachePath := c.cachePath(root)

	file, err := os.Open(cachePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entry CacheEntry
	if err := json.NewDecoder(file).Decode(&entry); err != nil {
		return nil, err
	}

	currentHash, err := c.computeHash(root)
	if err != nil {
		return nil, err
	}

	if entry.Hash != currentHash {
		return nil, fmt.Errorf("cache stale")
	}

	if time.Since(entry.Timestamp) > 24*time.Hour {
		return nil, fmt.Errorf("cache expired")
	}

	return entry.Graph, nil
}

func (c *Cache) Set(root string, g *Graph) error {
	hash, err := c.computeHash(root)
	if err != nil {
		return err
	}

	entry := CacheEntry{
		Hash:      hash,
		Timestamp: time.Now(),
		Graph:     g,
	}

	cachePath := c.cachePath(root)
	file, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(entry)
}

func (c *Cache) Clear(root string) error {
	return os.Remove(c.cachePath(root))
}
