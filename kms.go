package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/usegradient/gradient/internal/api"
)

const kmsUsage = `Usage: gradient kms <resource> <command> [args]

Resources:
  project   list | create <name> | get <id> | delete <id>
  branch    list <project_id|branch_id> | create <parent_id> <name> | get <id> | delete <id>
  secret    list <branch_id> | set <branch_id> <key> <value> | get <branch_id> <key> | delete <branch_id> <key>
  apply     <branch_id> <vm_id>
`

func runKMS(args []string, key string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, kmsUsage)
		return 1
	}
	client := api.NewClient(key)
	switch args[0] {
	case "project":
		return kmsProject(client, args[1:])
	case "branch":
		return kmsBranch(client, args[1:])
	case "secret":
		return kmsSecret(client, args[1:])
	case "apply":
		return kmsApply(client, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "gradient kms: unknown resource %q\n", args[0])
		fmt.Fprint(os.Stderr, kmsUsage)
		return 1
	}
}

func kmsProject(client *api.Client, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gradient kms project list | create <name> | get <id> | delete <id>")
		return 1
	}
	switch args[0] {
	case "list":
		resp, err := client.Get("/api/v1/kms/projects")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Print(string(resp.Data))
		return 0
	case "create":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: gradient kms project create <name>")
			return 1
		}
		body := map[string]string{"name": args[1]}
		resp, err := client.Post("/api/v1/kms/projects", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Print(string(resp.Data))
		return 0
	case "get":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: gradient kms project get <id>")
			return 1
		}
		resp, err := client.Get("/api/v1/kms/projects/" + url.PathEscape(args[1]))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Print(string(resp.Data))
		return 0
	case "delete":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: gradient kms project delete <id>")
			return 1
		}
		resp, err := client.Delete("/api/v1/kms/projects/" + url.PathEscape(args[1]))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Print(string(resp.Data))
		return 0
	default:
		fmt.Fprintf(os.Stderr, "gradient kms project: unknown command %q\n", args[0])
		return 1
	}
}

func kmsBranch(client *api.Client, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gradient kms branch list <id> | create <parent_id> <name> | get <id> | delete <id>")
		return 1
	}
	switch args[0] {
	case "list":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: gradient kms branch list <project_id|branch_id>")
			return 1
		}
		id := args[1]
		// Try project branches first, then branch children
		resp, err := client.Get("/api/v1/kms/projects/" + url.PathEscape(id) + "/branches")
		if err != nil {
			resp2, err2 := client.Get("/api/v1/kms/branches/" + url.PathEscape(id) + "/branches")
			if err2 != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return 1
			}
			fmt.Print(string(resp2.Data))
			return 0
		}
		fmt.Print(string(resp.Data))
		return 0
	case "create":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: gradient kms branch create <parent_branch_id> <name>")
			return 1
		}
		parentID, name := args[1], args[2]
		// Need project_id for POST /api/v1/kms/projects/:id/branches; API expects parent_branch_id and name
		// Get project from parent branch: we need project_id. API create branch is POST /projects/:id/branches with body { name, parent_branch_id }
		// So we need project_id. We can get it from branch get. So: get branch parent_id to know project_id.
		// Actually fleet-api: POST /api/v1/kms/projects/:id/branches with body { name, parent_branch_id }. So we need project_id. From the plan, "create <parent_id> <name>" - we don't have project_id. So we must get project_id from the parent branch. So first GET branch to get project_id? Or the API might accept parent_branch_id and resolve. Let me check fleet-api.
		// In fleet-api, POST is to /api/v1/kms/projects/:projectID/branches with body ParentBranchID, Name. So we need projectID. We can get it by GET /api/v1/kms/branches/:branch_id - but there's no "get branch by id" that returns project_id. There is GET /api/v1/kms/projects/:id/branches/:branch_id. So we'd need to know project id. Alternative: have user pass project_id: gradient kms branch create <project_id> <parent_branch_id> <name>. Plan says "create <parent_id> <name>". So we need to infer project. The only way is to get the branch and see which project it belongs to. API: GET /api/v1/kms/projects/:id/branches returns list. So we don't have "get branch by id" returning project. So for create we need either project_id in the API or we need an endpoint that accepts parent_branch_id. Looking at fleet-api again - POST projects/:id/branches - so we need project_id. So CLI could be gradient kms branch create <project_id> <parent_branch_id> <name>. Plan says create <parent_id> <name>. I'll use project_id from the first arg and parent_branch_id and name from next, or we need to ask. Actually re-read the fleet API code.
		// In handleKMS: POST /api/v1/kms/projects/:projectID/branches with body name, parent_branch_id. So we need projectID. So the CLI could be: gradient kms branch create <project_id> <parent_branch_id> <name>. But plan says create <parent_id> <name>. So we have two options: (1) change to create <project_id> <parent_id> <name>, or (2) first call GET projects, then for each project GET branches and find which project contains parent_branch_id, then create. (2) is complex. (1) is simpler. I'll do (1): gradient kms branch create <project_id> <parent_branch_id> <name> to match the API. Wait, the plan says "create <parent_id> <name>". So the minimal is parent_id and name. So we need to resolve project_id from parent_id. The only way without new API is to list all projects and then list all branches for each project and find which has parent_id. That's O(projects). I'll implement that: list projects, for each get branches, find branch with id == parent_id, then use that project_id to create.
		projectID, err := resolveProjectIDFromBranch(client, parentID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		body := map[string]string{"name": name, "parent_branch_id": parentID}
		resp, err := client.Post("/api/v1/kms/projects/"+url.PathEscape(projectID)+"/branches", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Print(string(resp.Data))
		return 0
	case "get":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: gradient kms branch get <id>")
			return 1
		}
		// Branch get: we need project_id and branch_id. API is GET /api/v1/kms/projects/:id/branches/:branch_id
		projectID, err := resolveProjectIDFromBranch(client, args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		resp, err := client.Get("/api/v1/kms/projects/" + url.PathEscape(projectID) + "/branches/" + url.PathEscape(args[1]))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Print(string(resp.Data))
		return 0
	case "delete":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: gradient kms branch delete <id>")
			return 1
		}
		projectID, err := resolveProjectIDFromBranch(client, args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		resp, err := client.Delete("/api/v1/kms/projects/" + url.PathEscape(projectID) + "/branches/" + url.PathEscape(args[1]))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Print(string(resp.Data))
		return 0
	default:
		fmt.Fprintf(os.Stderr, "gradient kms branch: unknown command %q\n", args[0])
		return 1
	}
}

func resolveProjectIDFromBranch(client *api.Client, branchID string) (string, error) {
	resp, err := client.Get("/api/v1/kms/projects")
	if err != nil {
		return "", err
	}
	var projects []struct {
		ID string `json:"id"`
	}
	if err := api.DataInto(resp, &projects); err != nil {
		return "", fmt.Errorf("parse projects: %w", err)
	}
	for _, p := range projects {
		// Check root stage branches
		r, err := client.Get("/api/v1/kms/projects/" + url.PathEscape(p.ID) + "/branches")
		if err != nil {
			continue
		}
		var stages []struct {
			ID string `json:"id"`
		}
		if err := api.DataInto(r, &stages); err != nil {
			continue
		}
		for _, s := range stages {
			if s.ID == branchID {
				return p.ID, nil
			}
			// Check child/forked branches under each stage
			cr, err := client.Get("/api/v1/kms/branches/" + url.PathEscape(s.ID) + "/branches")
			if err != nil {
				continue
			}
			var children []struct {
				ID string `json:"id"`
			}
			if err := api.DataInto(cr, &children); err != nil {
				continue
			}
			for _, c := range children {
				if c.ID == branchID {
					return p.ID, nil
				}
			}
		}
	}
	return "", fmt.Errorf("branch not found: %s", branchID)
}

func kmsSecret(client *api.Client, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gradient kms secret list <branch_id> | set <branch_id> <key> <value> | get <branch_id> <key> | delete <branch_id> <key>")
		return 1
	}
	switch args[0] {
	case "list":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: gradient kms secret list <branch_id>")
			return 1
		}
		projectID, err := resolveProjectIDFromBranch(client, args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		resp, err := client.Get("/api/v1/kms/projects/" + url.PathEscape(projectID) + "/branches/" + url.PathEscape(args[1]) + "/secrets")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Print(string(resp.Data))
		return 0
	case "set":
		if len(args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: gradient kms secret set <branch_id> <key> <value>")
			return 1
		}
		projectID, err := resolveProjectIDFromBranch(client, args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		body := map[string]string{args[2]: args[3]}
		resp, err := client.Put("/api/v1/kms/projects/"+url.PathEscape(projectID)+"/branches/"+url.PathEscape(args[1])+"/secrets", body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Print(string(resp.Data))
		return 0
	case "get":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: gradient kms secret get <branch_id> <key>")
			return 1
		}
		projectID, err := resolveProjectIDFromBranch(client, args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		resp, err := client.Get("/api/v1/kms/projects/" + url.PathEscape(projectID) + "/branches/" + url.PathEscape(args[1]) + "/secrets")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		var secrets map[string]string
		if err := api.DataInto(resp, &secrets); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		key := args[2]
		if v, ok := secrets[key]; ok {
			fmt.Println(v)
			return 0
		}
		fmt.Fprintf(os.Stderr, "Error: secret %q not found\n", key)
		return 1
	case "delete":
		// API has no DELETE single secret; report unsupported
		fmt.Fprintln(os.Stderr, "Error: deleting a single secret is not supported by the API. Use the dashboard or fleet CLI.")
		return 1
	default:
		fmt.Fprintf(os.Stderr, "gradient kms secret: unknown command %q\n", args[0])
		return 1
	}
}

func kmsApply(client *api.Client, args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: gradient kms apply <branch_id> <vm_id>")
		return 1
	}
	branchID, vmID := args[0], args[1]
	body := map[string]string{"vm_id": vmID}
	resp, err := client.Post("/api/v1/kms/branches/"+url.PathEscape(branchID)+"/apply", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Print(string(resp.Data))
	return 0
}
