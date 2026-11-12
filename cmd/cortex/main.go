package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mrvicadai/cortex/internal/beliefs"
	"github.com/mrvicadai/cortex/internal/notebook"
	"github.com/mrvicadai/cortex/internal/server"
)

const version = "0.1.0"

const helpText = `@mrvicadai/cortex v0.1.0
Unified Agent Context Architecture: normative beliefs and persistent episodic memory for AI coding agents.

USAGE:
  cortex                         Start MCP server over stdio transport
  cortex snippet                 Print zero-drift AGENTS.md dynamic router snippet
  cortex sync [path]             Update existing AGENTS.md router table in-place
  cortex instructions            Print agent best practices markdown
  cortex --db=<path>             Specify custom SQLite database path
  cortex --additional-beliefs=<dir> Specify custom user override directory for beliefs
  cortex --version, -v           Show version
  cortex --help, -h              Show this help message

TOOLS PROVIDED:
  Episodic Memory:
    cortex_context                            Load essential @context & recent events
    cortex_fact_put                           Create or CAS-replace a persistent fact
    cortex_fact_retire                        CAS-retire a fact with tombstone
    cortex_event_append                       Append immutable milestone outcome with evidence
    cortex_find                               Browse/search facts & events
    cortex_read                               Fetch complete unchunked records

  Normative Beliefs Engine:
    cortex_bootstrap                          Session bootstrap & topic router table
    cortex_get_belief                         Retrieve rules for topics (batch or single)
    cortex_beliefs_search                     Search rules with optional inline markdown

  Cross-Cutting:
    cortex_search                             Full-text search spanning beliefs & memory

HARNESS CONFIGURATION (npx zero-install):
  Claude Desktop, Cursor, or AGY:
  {
    "mcpServers": {
      "cortex": {
        "command": "npx",
        "args": ["-y", "@mrvicadai/cortex"]
      }
    }
  }
`

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, "cortex:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	// Fast subcommands before flag parsing
	if len(args) > 0 {
		switch args[0] {
		case "snippet", "--snippet":
			bStore, err := beliefs.NewStore("")
			if err != nil {
				return err
			}
			fmt.Fprintln(stdout, bStore.GetInlineSnippet())
			return nil
		case "sync", "--sync":
			var targets []string
			if len(args) > 1 {
				targets = []string{args[1]}
			} else {
				// Auto-detect AGENTS.md and CLAUDE.md in current directory
				if _, err := os.Stat("AGENTS.md"); err == nil {
					targets = append(targets, "AGENTS.md")
				}
				if _, err := os.Stat("CLAUDE.md"); err == nil {
					targets = append(targets, "CLAUDE.md")
				}
				if len(targets) == 0 {
					targets = []string{"AGENTS.md"}
				}
			}
			bStore, err := beliefs.NewStore("")
			if err != nil {
				return err
			}
			for _, targetPath := range targets {
				if err := syncFile(targetPath, bStore); err != nil {
					return err
				}
				fmt.Fprintf(stdout, "Successfully synced %s with current beliefs table.\n", targetPath)
			}
			return nil
		case "instructions", "--instructions":
			bStore, err := beliefs.NewStore("")
			if err != nil {
				return err
			}
			fmt.Fprintln(stdout, bStore.GetBootstrapPayload())
			return nil
		case "--version", "-v", "version":
			fmt.Fprintln(stdout, version)
			return nil
		case "--help", "-h", "help":
			fmt.Fprint(stdout, helpText)
			return nil
		}
	}

	flags := flag.NewFlagSet("cortex", flag.ContinueOnError)
	flags.SetOutput(stderr)

	dbFlag := flags.String("db", "", "SQLite database path (default: .cortex/cortex.sqlite or .fieldnotes/fieldnotes.sqlite)")
	overridesFlag := flags.String("additional-beliefs", "", "Directory containing custom markdown beliefs")
	vFlag := flags.Bool("v", false, "Show version")
	versionFlag := flags.Bool("version", false, "Show version")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *vFlag || *versionFlag {
		fmt.Fprintln(stdout, version)
		return nil
	}

	if flags.NArg() > 0 {
		return fmt.Errorf("unsupported argument '%s'", flags.Arg(0))
	}

	// 1. Resolve database path
	dbPath, err := resolveDatabasePath(*dbFlag)
	if err != nil {
		return err
	}

	// 2. Open Notebook Store
	nbStore, err := notebook.Open(dbPath)
	if err != nil {
		return fmt.Errorf("opening notebook store at %s: %w", dbPath, err)
	}
	defer nbStore.Close()

	// 3. Open Beliefs Store
	bStore, err := beliefs.NewStore(*overridesFlag)
	if err != nil {
		return fmt.Errorf("loading beliefs: %w", err)
	}

	// 4. Create and start MCP server
	srv := server.New(server.ServerOptions{
		Notebook: nbStore,
		Beliefs:  bStore,
	})

	return serve(srv, dbPath)
}

func resolveDatabasePath(custom string) (string, error) {
	if custom != "" {
		abs, err := filepath.Abs(custom)
		if err != nil {
			return "", err
		}
		if err := os.MkdirAll(filepath.Dir(abs), 0700); err != nil {
			return "", err
		}
		return abs, nil
	}

	// Automatic fallback: check if .fieldnotes/fieldnotes.sqlite exists
	if fi, err := os.Stat(filepath.Join(".fieldnotes", "fieldnotes.sqlite")); err == nil && !fi.IsDir() {
		return filepath.Abs(filepath.Join(".fieldnotes", "fieldnotes.sqlite"))
	}

	// Default to .cortex/cortex.sqlite in current directory
	target := filepath.Join(".cortex", "cortex.sqlite")
	abs, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0700); err != nil {
		return "", err
	}
	return abs, nil
}
