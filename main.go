package main

import (
	"fmt"
	"os"

	"github.com/usegradient/gradient/internal/config"
)

var Version = "dev"

const rootUsage = `Usage: gradient <command> [options] [args]

Commands:
  auth      Manage API key (login, logout, whoami, key)
  vm        VMs and projects (list, add, delete, info, up, down, resize, projects)
  kms       Secrets (project, branch, secret, apply)
  run       Run a command with env secrets injected (gradient run -- <cmd> [args])
  update    Update gradient to the latest version
  version   Print the current version

Use 'gradient <command>' for command-specific help.
`

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, rootUsage)
		return 0
	}
	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "auth":
		return runAuth(rest)
	case "update":
		return runUpdate(rest)
	case "version", "--version", "-v":
		fmt.Printf("gradient %s\n", Version)
		return 0
	case "help", "-h", "--help":
		fmt.Fprint(os.Stderr, rootUsage)
		return 0
	}

	hintUpdateIfAvailable()

	key, err := config.ReadAPIKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if key == "" {
		fmt.Fprintln(os.Stderr, "Error: not authenticated. Run 'gradient auth login' to set your API key.")
		return 1
	}

	switch cmd {
	case "vm":
		return runVM(rest, key)
	case "kms":
		return runKMS(rest, key)
	case "run":
		return runRun(rest, key)
	default:
		fmt.Fprintf(os.Stderr, "gradient: unknown command %q\n", cmd)
		fmt.Fprint(os.Stderr, rootUsage)
		return 1
	}
}
