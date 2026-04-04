# Command Reference

All commands are run through the Nix runner: `nix run .#<project> -- <command>`. When running directly (e.g., during development), the required environment variables must be set manually.

## `inframan init`

Initialize the Terraform workspace and pull remote state without applying changes.

**Usage:** `inframan init`

**Required env vars:** `INFRA_CONFIG_JSON`

**What it does:**
1. Copies the Terranix JSON config to `.inframan/<project>/terraform/config.tf.json`
2. Runs `terraform init` to connect to the backend and pull remote state

**Use case:** When you need Terraform state available (e.g., for `deploy` or `ssh`) but haven't run `infra` in this environment yet. Useful in CI or when switching machines.

```bash
nix run .#prod -- init
```

---

## `inframan infra`

Provision infrastructure using Terranix and Terraform.

**Usage:** `inframan infra`

**Required env vars:** `INFRA_CONFIG_JSON`

**What it does:**
1. Copies the Terranix JSON config to the Terraform workspace
2. Runs `terraform init`
3. Runs `terraform apply` (interactive -- Terraform shows a plan and prompts for confirmation)

**Note:** AWS credentials must be available, either via `terraform.tfvars` in the project's Terraform directory or via `TF_VAR_*` environment variables.

```bash
nix run .#prod -- infra
```

---

## `inframan deploy`

Deploy NixOS configuration to a provisioned instance using Colmena.

**Usage:** `inframan deploy`

**Required env vars:** `INFRA_CONFIG_JSON`, `NIXOS_MODULE_PATH`

**Optional env vars:** `SSH_KEY_PATH`, `SSH_CONFIG_PATH`, `SSH_PROXY_JUMP`, `USE_PRIVATE_IP`, `BUILD_ON_TARGET`, `TARGET_SYSTEM`

**What it does:**
1. Copies the Terranix config to the Terraform workspace (needed for remote backend init)
2. Runs `terraform output -json` to get the target IP
3. Generates an ephemeral `hive.nix` with the IP, SSH options, and your NixOS module injected
4. Runs `colmena apply --on target-node` to deploy

**Note:** The deploy command re-copies the Terranix config before reading Terraform output. This ensures `terraform init` works with remote backends even if the config file was generated in a different environment (e.g., Nix store path changes between machines).

```bash
nix run .#prod -- deploy
```

---

## `inframan destroy`

Destroy all infrastructure managed by Terraform.

**Usage:** `inframan destroy`

**What it does:**
1. Ensures Terraform is initialized (runs `terraform init` if `.terraform/` doesn't exist)
2. Runs `terraform destroy` (interactive -- prompts for confirmation)

```bash
nix run .#prod -- destroy
```

---

## `inframan ssh`

SSH to provisioned instances by project and instance name.

**Usage:** `inframan ssh [project[/instance]] [-- remote-command]`

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--user` | `-u` | `root` | SSH user |
| `--identity` | `-i` | (none) | Path to SSH identity file |
| `--list` | `-l` | `false` | List all available instances |

**Optional env vars:** `SSH_CONFIG_PATH`, `SSH_KEY_PATH`, `SSH_PROXY_JUMP`

**Examples:**

```bash
# List all available instances across all projects
inframan ssh --list

# Connect to a single-instance project
inframan ssh account1

# Connect to a specific instance in a multi-instance project
inframan ssh production/web-1

# Connect as a different user
inframan ssh account1 --user nixos

# Use a specific SSH key
inframan ssh account1 --identity ~/.ssh/id_ed25519

# Run a remote command
inframan ssh account1 -- hostname

# Pipe input to a remote command
echo "hello" | inframan ssh account1 -- cat
```

**SSH configuration priority:**
1. `SSH_CONFIG_PATH` (if set, custom SSH config takes full precedence; convenience options like `StrictHostKeyChecking=accept-new` are NOT added)
2. `--identity` flag
3. `SSH_KEY_PATH` env var
4. Default SSH behavior (no key specified)

**Implementation detail:** Uses `syscall.Exec()` to replace the inframan process with SSH, giving full terminal control.

---

## `inframan import`

Import existing resources into Terraform state.

**Usage:** `inframan import <address> <id>`

**Arguments (exactly 2 required):**
- `address` -- Terraform resource address (e.g., `aws_instance.main`)
- `id` -- Cloud provider resource ID (e.g., `i-1234567890abcdef0`)

**What it does:**
1. Ensures Terraform is initialized
2. Runs `terraform import <address> <id>`

```bash
nix run .#prod -- import aws_instance.main i-1234567890abcdef0
```

---

## `inframan output`

Show Terraform output values.

**Usage:** `inframan output [args...]`

**What it does:**
1. Ensures Terraform is initialized
2. Passes all arguments directly to `terraform output`

**Note:** Flag parsing is disabled for this command (`DisableFlagParsing: true`). All flags and arguments are passed through to Terraform unchanged.

```bash
# Show all outputs
nix run .#prod -- output

# Show outputs as JSON
nix run .#prod -- output -json

# Show a specific output
nix run .#prod -- output public_ip

# Show a specific output as raw value
nix run .#prod -- output -raw public_ip
```
