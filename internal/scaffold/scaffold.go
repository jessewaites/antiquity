// Package scaffold writes the folder structure and documents for a new case.
package scaffold

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"
)

//go:embed templates/*.tmpl
var files embed.FS

// Case is everything the wizard or the flags collect.
type Case struct {
	Slug        string
	Question    string
	Description string
	// Keys maps provider name → API key. Only non-empty values are written.
	Keys map[string]string
	Git  bool
}

// Providers the setup screen asks about, in display order.
var Providers = []struct{ Name, Label string }{
	{"anthropic", "Anthropic"},
	{"openai", "OpenAI"},
	{"typesafe", "TypeSafe (Jev)"},
	{"cloudflare", "Cloudflare"},
}

// Folders created inside the case, with the README line for each.
var Folders = []struct{ Name, Doc string }{
	{"sources", "One recipe per source: what it is, licence, how to fetch it, page images by reference, citation format, known traps. Start from `templates/source.md`. If the corpus already exists elsewhere on this machine, write its path here instead of copying it."},
	{"data", "Raw downloads, parquet, chunks and indexes. Gitignored. Suggested layout: `data/<source>/{raw,parquet,chunks,index}/`."},
	{"queries", "Retrieval configs: keyword regexes (with the known traps excluded), semantic queries, date windows. One file per query set, versioned."},
	{"judges", "Judge questions and extraction schemas, versioned. Start from `templates/judge.md`. The cheap judge runs over every candidate; the reader only sees survivors."},
	{"runs", "One folder per run: config snapshot, counts at each funnel stage (candidates → judged → read → verified), cost per model role, log."},
	{"candidates", "One row per candidate with status: new → sorted → read → verified / rejected. Rejects keep a reason."},
	{"evidence", "Page images for every find and every instructive reject: `<date>_<place>_<what>_<archive-ref>_{full,crop}.jpg`. Rejects get a `REJECTED_` prefix."},
	{"findings", "`FINDING_<date>_<slug>.md`, from `templates/FINDING.md`. Status candidate until read on the page image."},
	{"theories", "Hypotheses, one file each, with what would confirm or refute each."},
	{"catalogues", "The reference catalogues used for novelty checks, and for each: how it was checked (full read vs search-inside) and what was found."},
	{"controls", "Known events the pipeline must re-find. One file per control; per-run recall results. A negative result without a control is not a result."},
	{"todo", "Small actionable items, one file each, with progress notes."},
	{"drafts", "Outreach notes to specialists, plain text. Never sent by an agent."},
	{"templates", "Starting points for findings, source recipes and judge questions."},
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify turns free text into a folder-safe name.
func Slugify(s string) string {
	s = slugRe.ReplaceAllString(strings.ToLower(strings.TrimSpace(s)), "-")
	return strings.Trim(s, "-")
}

// ErrExists is returned when the target directory already has content.
var ErrExists = errors.New("directory already exists and is not empty")

// Write creates the case under parent/<slug>. It returns the case path.
func Write(parent string, c Case) (string, error) {
	if c.Slug == "" {
		return "", errors.New("case name is empty")
	}
	dir := filepath.Join(parent, c.Slug)
	if entries, err := os.ReadDir(dir); err == nil && len(entries) > 0 {
		return dir, fmt.Errorf("%w: %s", ErrExists, dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return dir, err
	}

	data := map[string]any{
		"Slug":        c.Slug,
		"Question":    strings.TrimSpace(c.Question),
		"Description": strings.TrimSpace(c.Description),
		"Date":        time.Now().Format("2006-01-02"),
		"Keys":        c.keyNames(),
		"KeyValues":   c.nonEmptyKeys(),
	}

	render := func(tmpl, dest string, mode os.FileMode) error {
		t, err := template.New(tmpl).Funcs(template.FuncMap{"join": strings.Join}).ParseFS(files, "templates/"+tmpl)
		if err != nil {
			return err
		}
		var sb strings.Builder
		if err := t.Execute(&sb, data); err != nil {
			return fmt.Errorf("%s: %w", tmpl, err)
		}
		path := filepath.Join(dir, dest)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		return os.WriteFile(path, []byte(sb.String()), mode)
	}

	docs := []struct{ tmpl, dest string }{
		{"AGENTS.md.tmpl", "AGENTS.md"},
		{"CASE.md.tmpl", "CASE.md"},
		{"HANDOFF.md.tmpl", "HANDOFF.md"},
		{"NARRATIVE.md.tmpl", "NARRATIVE.md"},
		{"lessons.md.tmpl", "lessons.md"},
		{"gitignore.tmpl", ".gitignore"},
		{"keys.example.yml.tmpl", "keys.example.yml"},
		{"FINDING.md.tmpl", "templates/FINDING.md"},
		{"source.md.tmpl", "templates/source.md"},
		{"judge.md.tmpl", "templates/judge.md"},
	}
	for _, d := range docs {
		if err := render(d.tmpl, d.dest, 0o644); err != nil {
			return dir, err
		}
	}
	if len(c.nonEmptyKeys()) > 0 {
		if err := render("keys.yml.tmpl", "keys.yml", 0o600); err != nil {
			return dir, err
		}
	}
	for _, f := range Folders {
		readme := filepath.Join(dir, f.Name, "README.md")
		if err := os.MkdirAll(filepath.Dir(readme), 0o755); err != nil {
			return dir, err
		}
		if err := os.WriteFile(readme, []byte("# "+f.Name+"\n\n"+f.Doc+"\n"), 0o644); err != nil {
			return dir, err
		}
	}
	if c.Git {
		if err := initGit(dir, render); err != nil {
			return dir, err
		}
	}
	return dir, nil
}

func initGit(dir string, render func(tmpl, dest string, mode os.FileMode) error) error {
	if _, err := exec.LookPath("git"); err != nil {
		return nil // no git: silently skip
	}
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init: %s", strings.TrimSpace(string(out)))
	}
	return render("pre-commit.tmpl", filepath.Join(".git", "hooks", "pre-commit"), 0o755)
}

func (c Case) nonEmptyKeys() map[string]string {
	out := map[string]string{}
	for k, v := range c.Keys {
		if strings.TrimSpace(v) != "" {
			out[k] = strings.TrimSpace(v)
		}
	}
	return out
}

func (c Case) keyNames() []string {
	var names []string
	for k := range c.nonEmptyKeys() {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// Tree renders the structure that Write will create, for the confirm screen.
func Tree(c Case) string {
	var sb strings.Builder
	sb.WriteString(c.Slug + "/\n")
	top := []string{"AGENTS.md", "CASE.md", "HANDOFF.md", "NARRATIVE.md", "lessons.md", "keys.example.yml"}
	if len(c.nonEmptyKeys()) > 0 {
		top = append(top, "keys.yml  (chmod 600, gitignored)")
	}
	for _, t := range top {
		sb.WriteString("  " + t + "\n")
	}
	var names []string
	for _, f := range Folders {
		names = append(names, f.Name+"/")
	}
	line := "  "
	for _, n := range names {
		if len(line)+len(n)+1 > 60 {
			sb.WriteString(line + "\n")
			line = "  "
		}
		line += n + " "
	}
	sb.WriteString(strings.TrimRight(line, " ") + "\n")
	return sb.String()
}
