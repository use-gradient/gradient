package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/usegradient/gradient/internal/api"
)

const vmUsage = `Usage: gradient vm <command> [options] [args]

Commands:
  list                    List all VMs
  add <name>              Create a VM (requires --project)
  delete <name>           Delete a VM
  info <name>             Show VM details
  up <name>               Start a VM
  down <name>             Stop a VM
  resize <name>           Resize VM (--balloon, --memory, --cpus)
  projects                List projects
  projects <name>         List VMs in project
  projects delete <name>  Delete a project and its VMs
`

func runVM(args []string, key string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, vmUsage)
		return 1
	}
	client := api.NewClient(key)
	switch args[0] {
	case "list":
		return vmList(client, args[1:])
	case "add":
		return vmAdd(client, args[1:])
	case "delete":
		return vmDelete(client, args[1:])
	case "info":
		return vmInfo(client, args[1:])
	case "up":
		return vmUp(client, args[1:])
	case "down":
		return vmDown(client, args[1:])
	case "resize":
		return vmResize(client, args[1:])
	case "projects":
		return vmProjects(client, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "gradient vm: unknown command %q\n", args[0])
		fmt.Fprint(os.Stderr, vmUsage)
		return 1
	}
}

func vmList(client *api.Client, args []string) int {
	resp, err := client.Get("/api/v1/vm")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Print(string(resp.Data))
	return 0
}

func vmAdd(client *api.Client, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gradient vm add <name> --project <project> [--cpus N] [--memory SIZE] [--disk SIZE] [--repo URL]")
		return 1
	}
	name := args[0]
	var project, cpus, memory, disk, repo string
	rest := args[1:]
	for i := 0; i < len(rest); i++ {
		switch rest[i] {
		case "--project":
			if i+1 < len(rest) {
				project = rest[i+1]
				i++
			}
		case "--cpus":
			if i+1 < len(rest) {
				cpus = rest[i+1]
				i++
			}
		case "--memory":
			if i+1 < len(rest) {
				memory = rest[i+1]
				i++
			}
		case "--disk":
			if i+1 < len(rest) {
				disk = rest[i+1]
				i++
			}
		case "--repo":
			if i+1 < len(rest) {
				repo = rest[i+1]
				i++
			}
		}
	}
	if project == "" {
		fmt.Fprintln(os.Stderr, "Error: --project is required")
		return 1
	}
	body := map[string]string{"name": name, "project": project}
	if cpus != "" {
		body["cpus"] = cpus
	}
	if memory != "" {
		body["memory"] = memory
	}
	if disk != "" {
		body["disk"] = disk
	}
	if repo != "" {
		body["repo"] = repo
	}
	resp, err := client.Post("/api/v1/vm", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Print(string(resp.Data))
	return 0
}

func vmDelete(client *api.Client, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gradient vm delete <name>")
		return 1
	}
	name := args[0]
	resp, err := client.Delete("/api/v1/vm/" + pathEscape(name))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Print(string(resp.Data))
	return 0
}

func vmInfo(client *api.Client, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gradient vm info <name>")
		return 1
	}
	name := args[0]
	resp, err := client.Get("/api/v1/vm/" + pathEscape(name))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Print(string(resp.Data))
	return 0
}

func vmUp(client *api.Client, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gradient vm up <name>")
		return 1
	}
	name := args[0]
	resp, err := client.Post("/api/v1/vm/"+pathEscape(name)+"/up", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Print(string(resp.Data))
	return 0
}

func vmDown(client *api.Client, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gradient vm down <name>")
		return 1
	}
	name := args[0]
	resp, err := client.Post("/api/v1/vm/"+pathEscape(name)+"/down", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Print(string(resp.Data))
	return 0
}

func vmResize(client *api.Client, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: gradient vm resize <name> [--balloon SIZE] [--memory SIZE] [--cpus N]")
		return 1
	}
	name := args[0]
	var balloon, memory, cpus string
	rest := args[1:]
	for i := 0; i < len(rest); i++ {
		switch rest[i] {
		case "--balloon":
			if i+1 < len(rest) {
				balloon = rest[i+1]
				i++
			}
		case "--memory":
			if i+1 < len(rest) {
				memory = rest[i+1]
				i++
			}
		case "--cpus":
			if i+1 < len(rest) {
				cpus = rest[i+1]
				i++
			}
		}
	}
	body := map[string]string{}
	if balloon != "" {
		body["balloon"] = balloon
	}
	if memory != "" {
		body["memory"] = memory
	}
	if cpus != "" {
		body["cpus"] = cpus
	}
	resp, err := client.Post("/api/v1/vm/"+pathEscape(name)+"/resize", body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Print(string(resp.Data))
	return 0
}

func vmProjects(client *api.Client, args []string) int {
	if len(args) == 0 {
		resp, err := client.Get("/api/v1/vm/projects")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Print(string(resp.Data))
		return 0
	}
	if args[0] == "delete" {
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: gradient vm projects delete <project-name>")
			return 1
		}
		projectName := args[1]
		resp, err := client.Delete("/api/v1/vm/projects/" + pathEscape(projectName))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		fmt.Print(string(resp.Data))
		return 0
	}
	projectName := args[0]
	resp, err := client.Get("/api/v1/vm/projects/" + pathEscape(projectName))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Print(string(resp.Data))
	return 0
}

func pathEscape(s string) string {
	return url.PathEscape(s)
}
