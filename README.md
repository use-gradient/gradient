# Gradient CLI

Command-line client for [Gradient](https://usegradient.dev) — manage VMs, secrets, and inject environment variables into any process.

## Installation

```sh
curl -fsSL https://raw.githubusercontent.com/use-gradient/gradient/main/install.sh | sudo sh
```

This detects your OS and architecture, downloads the binary, and installs it to `/usr/local/bin/gradient`.

To install to a different location:

```sh
GRADIENT_BIN=~/.local/bin/gradient curl -fsSL https://raw.githubusercontent.com/use-gradient/gradient/main/install.sh | sudo sh
```

## Authentication

All commands (except `auth`) require an API key. Get yours from the [Gradient dashboard](https://console.usegradient.dev), then:

```sh
gradient auth login
# Prompts for your API key and stores it locally
```


| Command                | Description                      |
| ---------------------- | -------------------------------- |
| `gradient auth login`  | Store your API key               |
| `gradient auth logout` | Remove stored credentials        |
| `gradient auth whoami` | Verify your key and check access |
| `gradient auth key`    | Print the stored API key         |


Credentials are saved to `~/.config/gradient/credentials`.

## VMs

```sh
gradient vm <command> [options]
```


| Command                                      | Description                                     |
| -------------------------------------------- | ----------------------------------------------- |
| `gradient vm list`                           | List all your VMs                               |
| `gradient vm add <name> --project <project>` | Create a VM in a project                        |
| `gradient vm delete <name>`                  | Delete a VM                                     |
| `gradient vm info <name>`                    | Show VM details (status, CPU, memory, ULA)      |
| `gradient vm up <name>`                      | Start a stopped VM                              |
| `gradient vm down <name>`                    | Stop a running VM                               |
| `gradient vm resize <name> [flags]`          | Resize a VM (`--cpus`, `--memory`, `--balloon`) |
| `gradient vm projects`                       | List all projects                               |
| `gradient vm projects <name>`                | List VMs in a project                           |
| `gradient vm projects delete <name>`         | Delete a project and all its VMs                |


`vm add` also accepts `--cpus`, `--memory`, `--disk`, and `--repo`.

## Secrets (KMS)

Gradient includes a built-in secrets manager with three default stages per project: **dev**, **staging**, and **prod**. Stages can be forked into child branches for per-developer or per-feature overrides.

### Projects


| Command                              | Description                                                 |
| ------------------------------------ | ----------------------------------------------------------- |
| `gradient kms project list`          | List all KMS projects                                       |
| `gradient kms project create <name>` | Create a new project (auto-creates dev/staging/prod stages) |
| `gradient kms project get <id>`      | Get project details                                         |
| `gradient kms project delete <id>`   | Delete a project                                            |


### Branches


| Command                                                | Description                                               |
| ------------------------------------------------------ | --------------------------------------------------------- |
| `gradient kms branch list <project_id|branch_id>`      | List stages for a project, or child branches for a branch |
| `gradient kms branch create <parent_branch_id> <name>` | Fork a new branch from a parent                           |
| `gradient kms branch get <id>`                         | Get branch details                                        |
| `gradient kms branch delete <id>`                      | Delete a branch                                           |


### Secrets


| Command                                             | Description                                             |
| --------------------------------------------------- | ------------------------------------------------------- |
| `gradient kms secret list <branch_id>`              | List all secrets on a branch                            |
| `gradient kms secret set <branch_id> <key> <value>` | Set a secret (propagates to sibling root stages if new) |
| `gradient kms secret get <branch_id> <key>`         | Get a single secret value                               |


### Apply

Push secrets from a branch directly to a VM:

```sh
gradient kms apply <branch_id> <vm_id>
```

## Environment Injection

Run any command with secrets from a KMS branch injected as environment variables — similar to `doppler run`:

```sh
gradient run -- <command> [args...]
```

On first run, if no `.gradient.yaml` exists in the current directory, Gradient will interactively prompt you to select a KMS project and branch, then save the selection for future use:

```yaml
# .gradient.yaml
project_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
branch_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
```

Example:

```sh
gradient run -- node server.js
gradient run -- python manage.py runserver
gradient run -- env  # prints all env vars including injected secrets
```

## Updating

Gradient checks for new releases once per day and prints a hint when one is available. To update:

```sh
gradient update
```

To check your current version:

```sh
gradient version
```

## Security

Secrets fetched via the CLI are encrypted in transit using AES-256-GCM with a key derived from your API key. This provides an additional layer of encryption on top of HTTPS, so secret values are never transmitted in plaintext even within the TLS tunnel.

## Configuration

| Item                  | Location                                     |
| --------------------- | -------------------------------------------- |
| API key               | `~/.config/gradient/credentials`             |
| Project config        | `.gradient.yaml` (in your working directory) |
| API base URL override | `GRADIENT_API_URL` environment variable       |

## Building from source

```sh
git clone https://github.com/use-gradient/gradient.git
cd gradient
go build -ldflags "-X main.Version=dev" -o gradient .
```

