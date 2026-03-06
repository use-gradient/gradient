package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/usegradient/gradient/internal/api"
	"github.com/usegradient/gradient/internal/config"
)

func runRun(args []string, key string) int {
	// Parse "gradient run -- <command> [args...]"
	var cmdArgs []string
	for i, a := range args {
		if a == "--" {
			if i+1 < len(args) {
				cmdArgs = args[i+1:]
			}
			break
		}
	}
	if len(cmdArgs) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gradient run -- <command> [args...]")
		fmt.Fprintln(os.Stderr, "Runs the command with env secrets from your KMS branch injected.")
		return 1
	}

	client := api.NewClient(key)
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	cfg, err := config.ReadProjectConfig(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if cfg == nil {
		// Interactive: select project and branch, write .gradient.yaml
		cfg, err = promptProjectAndBranch(client, cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		if err := config.WriteProjectConfig(cwd, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Fprintln(os.Stderr, "Created .gradient.yaml for future runs. You can edit it to change project or branch.")
	}

	// Fetch secrets
	resp, err := client.Get("/api/v1/kms/projects/" + url.PathEscape(cfg.ProjectID) + "/branches/" + url.PathEscape(cfg.BranchID) + "/secrets")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	var secrets map[string]string
	if err := api.DataInto(resp, &secrets); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Build env: current env + secrets (secrets override)
	env := os.Environ()
	for k, v := range secrets {
		env = append(env, k+"="+v)
	}

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = cwd
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func promptProjectAndBranch(client *api.Client, cwd string) (*config.ProjectConfig, error) {
	// List KMS projects
	resp, err := client.Get("/api/v1/kms/projects")
	if err != nil {
		return nil, err
	}
	var projects []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := api.DataInto(resp, &projects); err != nil {
		return nil, fmt.Errorf("parse projects: %w", err)
	}
	if len(projects) == 0 {
		return nil, fmt.Errorf("no KMS projects found; create one with gradient kms project create <name>")
	}

	fmt.Fprintln(os.Stderr, "Select a KMS project:")
	for i, p := range projects {
		fmt.Fprintf(os.Stderr, "  %d) %s\n", i+1, p.Name)
	}
	fmt.Fprint(os.Stderr, "Name or number: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return nil, fmt.Errorf("reading input")
	}
	input := strings.TrimSpace(scanner.Text())
	idx := -1
	if n, err := fmt.Sscanf(input, "%d", &idx); n == 1 && err == nil && idx >= 1 && idx <= len(projects) {
		idx-- // convert to 0-based
	} else {
		idx = -1
		for i, p := range projects {
			if strings.EqualFold(p.Name, input) {
				idx = i
				break
			}
		}
	}
	if idx < 0 {
		return nil, fmt.Errorf("invalid selection: %q", input)
	}
	projectID := projects[idx].ID

	// List branches for this project
	resp2, err := client.Get("/api/v1/kms/projects/" + url.PathEscape(projectID) + "/branches")
	if err != nil {
		return nil, err
	}
	var branches []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Stage string `json:"stage"`
	}
	if err := api.DataInto(resp2, &branches); err != nil {
		return nil, fmt.Errorf("parse branches: %w", err)
	}
	if len(branches) == 0 {
		return nil, fmt.Errorf("no branches in project")
	}

	fmt.Fprintln(os.Stderr, "Select a branch:")
	for i, b := range branches {
		stage := b.Stage
		if stage == "" {
			stage = b.Name
		}
		fmt.Fprintf(os.Stderr, "  %d) %s\n", i+1, stage)
	}
	fmt.Fprint(os.Stderr, "Name or number: ")
	if !scanner.Scan() {
		return nil, fmt.Errorf("reading input")
	}
	binput := strings.TrimSpace(scanner.Text())
	bidx := -1
	if n, err := fmt.Sscanf(binput, "%d", &bidx); n == 1 && err == nil && bidx >= 1 && bidx <= len(branches) {
		bidx--
	} else {
		bidx = -1
		for i, b := range branches {
			label := b.Stage
			if label == "" {
				label = b.Name
			}
			if strings.EqualFold(label, binput) {
				bidx = i
				break
			}
		}
	}
	if bidx < 0 {
		return nil, fmt.Errorf("invalid selection: %q", binput)
	}
	branchID := branches[bidx].ID

	return &config.ProjectConfig{ProjectID: projectID, BranchID: branchID}, nil
}
