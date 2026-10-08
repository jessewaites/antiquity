package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/jessewaites/antiquity/internal/intro"
	"github.com/jessewaites/antiquity/internal/scaffold"
	"github.com/jessewaites/antiquity/internal/wizard"
)

const tDoneMs = 5300

const usage = `Antiquity — AI-assisted historical investigation

Usage:
  antiquity new [case-name] [flags]   open a new case
      --question "..."      the one-line question
      --description "..."   more context for the agent
      --yes                 skip the wizard (needs a name and --question)
      --no-intro            skip the title screen
      --no-git              do not git init the case
      --dir PATH            parent directory (default: current)
  antiquity intro [--still] [--snapshot] [--at MS] [--case NAME]
                                      play the title screen on its own
`

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		fmt.Print(usage)
		return
	}
	switch os.Args[1] {
	case "new":
		cmdNew(os.Args[2:])
	case "intro":
		cmdIntro(os.Args[2:])
	default:
		fmt.Fprintln(os.Stderr, "Unknown command. Try: antiquity new <case-name>")
		os.Exit(2)
	}
}

func cmdNew(args []string) {
	flags := flag.NewFlagSet("new", flag.ExitOnError)
	question := flags.String("question", "", "the one-line question")
	description := flags.String("description", "", "more context for the agent")
	yes := flags.Bool("yes", false, "skip the wizard")
	noIntro := flags.Bool("no-intro", false, "skip the title screen")
	noGit := flags.Bool("no-git", false, "do not git init")
	dir := flags.String("dir", ".", "parent directory")
	flags.Usage = func() { fmt.Print(usage) }
	// Allow the name before or after flags.
	var name string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name, args = args[0], args[1:]
	}
	flags.Parse(args)
	if name == "" && flags.NArg() > 0 {
		name = flags.Arg(0)
	}

	parent, err := filepath.Abs(*dir)
	if err != nil {
		fail(err)
	}
	c := scaffold.Case{
		Slug:        scaffold.Slugify(name),
		Question:    *question,
		Description: *description,
		Git:         !*noGit,
	}

	if *yes {
		if c.Slug == "" || strings.TrimSpace(c.Question) == "" {
			fail(errors.New("--yes needs a case name and --question"))
		}
	} else {
		if !*noIntro {
			im := run(intro.New(c.Slug, false)).(intro.Model)
			if !im.Proceed {
				return
			}
		}
		wm := run(wizard.New(shorten(parent), c)).(wizard.Model)
		if !wm.Confirmed {
			return
		}
		c = wm.Case
	}

	path, err := scaffold.Write(parent, c)
	if err != nil {
		fail(err)
	}
	fmt.Printf("Created %s\n\n%s\n", shorten(path), scaffold.Tree(c))
	fmt.Printf("Next:\n  cd %s\n  open your coding agent; it reads AGENTS.md first.\n", shorten(path))
	if len(c.Keys) == 0 || allBlank(c.Keys) {
		fmt.Println("  copy keys.example.yml to keys.yml and fill in what you have.")
	}
}

func cmdIntro(args []string) {
	flags := flag.NewFlagSet("intro", flag.ExitOnError)
	still := flags.Bool("still", false, "start with the animation finished")
	snapshot := flags.Bool("snapshot", false, "print one frame without entering the TUI")
	at := flags.Int("at", tDoneMs, "snapshot time in ms")
	caseName := flags.String("case", "test-mystery", "case name to show")
	flags.Parse(args)
	m := intro.New(*caseName, *still)
	if *snapshot {
		fmt.Println(m.Snapshot(80, 24, *at))
		return
	}
	run(m)
}

func run(m tea.Model) tea.Model {
	out, err := tea.NewProgram(m).Run()
	if err != nil {
		fail(fmt.Errorf("could not open the terminal: %w", err))
	}
	return out
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "antiquity:", err)
	os.Exit(1)
}

func shorten(path string) string {
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

func allBlank(keys map[string]string) bool {
	for _, v := range keys {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}
