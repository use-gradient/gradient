package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/usegradient/gradient/internal/api"
	"github.com/usegradient/gradient/internal/config"
)

const authUsage = `Usage: gradient auth <command>

Commands:
  login   Store your API key (prompts if not set)
  logout  Remove stored API key
  whoami  Verify key and show current user
  key     Print stored API key
`

func runAuth(args []string) int {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, authUsage)
		return 1
	}
	switch args[0] {
	case "login":
		return runAuthLogin(args[1:])
	case "logout":
		return runAuthLogout(args[1:])
	case "whoami":
		return runAuthWhoami(args[1:])
	case "key":
		return runAuthKey(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "gradient auth: unknown command %q\n", args[0])
		fmt.Fprint(os.Stderr, authUsage)
		return 1
	}
}

func runAuthLogin(args []string) int {
	key, err := config.ReadAPIKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if key != "" {
		fmt.Fprintln(os.Stderr, "Already logged in. Use 'gradient auth logout' first to replace the key.")
		return 0
	}
	fmt.Fprint(os.Stderr, "API key: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		fmt.Fprintln(os.Stderr, "Error: reading input")
		return 1
	}
	key = strings.TrimSpace(scanner.Text())
	if key == "" {
		fmt.Fprintln(os.Stderr, "Error: API key cannot be empty")
		return 1
	}
	if err := config.WriteAPIKey(key); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Fprintln(os.Stderr, "API key stored.")

	// Register device key for E2E encryption if not already present
	_, deviceID, err := config.ReadDeviceKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not read device key: %v\n", err)
		return 0
	}
	if deviceID != "" {
		return 0
	}
	priv, pub, err := config.GenerateDeviceKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not generate device key: %v\n", err)
		return 0
	}
	client := api.NewClient(key, "", nil)
	resp, err := client.Post("/api/v1/auth/devices", map[string]string{
		"name":       "cli",
		"public_key": base64.StdEncoding.EncodeToString(pub),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not register device key: %v\n", err)
		return 0
	}
	deviceKeyID := extractDeviceID(resp.Data)
	if deviceKeyID == "" {
		fmt.Fprintf(os.Stderr, "Warning: invalid device registration response\n")
		return 0
	}
	if err := config.WriteDeviceKey(priv, deviceKeyID); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not save device key: %v\n", err)
		return 0
	}
	fmt.Fprintln(os.Stderr, "Device key registered for end-to-end encryption.")
	return 0
}

func extractDeviceID(raw json.RawMessage) string {
	// Case 1: data is a JSON object with "id" field
	var obj struct {
		ID string `json:"id"`
	}
	if json.Unmarshal(raw, &obj) == nil && obj.ID != "" {
		return strings.TrimSpace(obj.ID)
	}
	// Case 2: data is a JSON string (server wrapped it); try parsing that string as JSON object
	var s string
	if json.Unmarshal(raw, &s) == nil && s != "" {
		s = strings.SplitN(s, "\n", 2)[0] // drop any trailing status lines
		if json.Unmarshal([]byte(s), &obj) == nil && obj.ID != "" {
			return strings.TrimSpace(obj.ID)
		}
		s = strings.TrimSpace(s)
		if len(s) > 0 && s[0] != '{' {
			return s
		}
	}
	return ""
}

func runAuthLogout(args []string) int {
	if err := config.DeleteCredentials(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	_ = config.DeleteDeviceKey()
	fmt.Fprintln(os.Stderr, "Logged out.")
	return 0
}

func runAuthWhoami(args []string) int {
	key, err := config.ReadAPIKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if key == "" {
		fmt.Fprintln(os.Stderr, "Error: not authenticated. Run 'gradient auth login' to set your API key.")
		return 1
	}
	priv, deviceID, _ := config.ReadDeviceKey()
	client := api.NewClient(key, deviceID, priv)
	resp, err := client.Get("/api/v1/vm/projects")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	// Response is OK and we have access; we don't have a "username" in the API, so just confirm auth.
	fmt.Println("Authenticated. Your API key has access to fleet.")
	if len(resp.Data) > 0 && string(resp.Data) != "null" {
		fmt.Println("Projects: (list available via gradient vm projects)")
	}
	return 0
}

func runAuthKey(args []string) int {
	key, err := config.ReadAPIKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if key == "" {
		fmt.Fprintln(os.Stderr, "Error: not authenticated. Run 'gradient auth login' to set your API key.")
		return 1
	}
	fmt.Println(key)
	return 0
}
