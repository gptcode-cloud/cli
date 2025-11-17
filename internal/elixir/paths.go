package elixir

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Project struct {
	Root       string
	AppName    string
	ModuleBase string
}

func Detect(root string) (*Project, error) {
	if root == "" {
		r, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("getcwd: %w", err)
		}
		root = r
	}

	cur := root
	for {
		mixPath := filepath.Join(cur, "mix.exs")
		if _, err := os.Stat(mixPath); err == nil {
			return parseMixFile(cur, mixPath)
		}

		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}

	return nil, fmt.Errorf("no mix.exs found from %s upward", root)
}

func parseMixFile(root, mixPath string) (*Project, error) {
	data, err := os.ReadFile(mixPath)
	if err != nil {
		return nil, fmt.Errorf("read mix.exs: %w", err)
	}
	src := string(data)

	reApp := regexp.MustCompile(`app:\\s*:(\\w+)`)
	app := "app"
	if m := reApp.FindStringSubmatch(src); len(m) >= 2 {
		app = m[1]
	}

	reMod := regexp.MustCompile(`mod:\\s*{\\s*([\\w\\.]+)`)
	moduleBase := moduleNamespace(app)
	if m := reMod.FindStringSubmatch(src); len(m) >= 2 {
		moduleBase = strings.Split(m[1], ".")[0]
	}

	return &Project{
		Root:       root,
		AppName:    app,
		ModuleBase: moduleBase,
	}, nil
}

func moduleNamespace(app string) string {
	parts := strings.Split(app, "_")
	for i, p := range parts {
		if len(p) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}

func SlugForDescription(desc string) string {
	desc = strings.ToLower(desc)
	re := regexp.MustCompile(`[^a-z0-9\\s_]+`)
	desc = re.ReplaceAllString(desc, " ")
	parts := strings.Fields(desc)
	if len(parts) == 0 {
		return "feature"
	}
	if len(parts) == 1 {
		return parts[0]
	}

	drop := map[string]bool{
		"calculate": true,
		"compute":   true,
		"manage":    true,
		"handle":    true,
		"process":   true,
		"support":   true,
		"list":      true,
		"create":    true,
		"update":    true,
		"delete":    true,
	}
	var kept []string
	for _, p := range parts {
		if drop[p] {
			continue
		}
		kept = append(kept, p)
	}
	if len(kept) == 0 {
		kept = parts
	}
	if len(kept) > 2 {
		kept = kept[:2]
	}
	return strings.Join(kept, "_")
}

func ModuleNameForSlug(slug string) string {
	if slug == "" {
		return "Feature"
	}
	parts := strings.Split(slug, "_")
	for i, p := range parts {
		if len(p) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}

func PathsForSlug(p *Project, slug string) (testPath, implPath string) {
	if slug == "" {
		slug = "feature"
	}
	testPath = filepath.Join("test", p.AppName, slug+"_test.exs")
	implPath = filepath.Join("lib", p.AppName, slug+".ex")
	return
}

