# Configuration Reference

Inframan is configured through environment variables. In normal usage, these are set automatically by the `lib.mkRunner` Nix function. When running the Go binary directly (e.g., during development), you must set them manually.

## `lib.mkRunner` Parameters

These parameters are passed to `inframan.lib.mkRunner` in your project's `flake.nix`:

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `system` | string | Yes | -- | Build system architecture (e.g., `"x86_64-linux"`, `"aarch64-darwin"`) |
| `infraConfig` | path | Yes | -- | Path to Terranix infrastructure `.nix` file |
| `machineConfig` | path | Yes | -- | Path to NixOS machine configuration `.nix` file |
| `projectName` | string | No | `"default"` | Project name, used for `.inframan/<projectName>/` isolation |
| `targetSystem` | string | No | `"x86_64-linux"` | Target machine's NixOS architecture (e.g., `"aarch64-linux"`) |
| `sshKeyPath` | string/null | No | `null` | Path to SSH private key for deployment and SSH access |
| `sshConfigPath` | string/null | No | `null` | Path to SSH config file (useful for multi-user setups) |
| `sshProxyJump` | string/null | No | `null` | SSH proxy jump host (e.g., `"user@bastion:22"`) |
| `buildOnTarget` | bool | No | `true` | Build NixOS config on remote target vs locally |
| `usePrivateIp` | bool | No | `false` | Prefer `private_ip` from Terraform output |

### Example

```nix
inframan.lib.mkRunner {
  system = "x86_64-linux";
  infraConfig = ./infrastructure.nix;
  machineConfig = ./machine.nix;
  projectName = "production";
  targetSystem = "aarch64-linux";
  sshKeyPath = "./keys/deploy_key";
  sshProxyJump = "admin@bastion.example.com:22";
  buildOnTarget = false;
  usePrivateIp = true;
}
```

## Environment Variables

These are set by `lib.mkRunner` and read by the Go binary:

| Variable | Set By | Default | Used By | Description |
|----------|--------|---------|---------|-------------|
| `INFRA_CONFIG_JSON` | mkRunner | (required) | init, infra, deploy | Nix store path to Terranix-generated JSON |
| `NIXOS_MODULE_PATH` | mkRunner | (required) | deploy | Nix store path to NixOS module |
| `PROJECT_NAME` | mkRunner | `"default"` | all | Project name for directory isolation |
| `TARGET_SYSTEM` | mkRunner | `"x86_64-linux"` | deploy | Target architecture in generated `hive.nix` |
| `SSH_KEY_PATH` | mkRunner | (empty) | deploy, ssh | SSH private key path |
| `SSH_CONFIG_PATH` | mkRunner | (empty) | deploy, ssh | SSH config file path |
| `SSH_PROXY_JUMP` | mkRunner | (empty) | deploy, ssh | Bastion/jump host for SSH |
| `USE_PRIVATE_IP` | mkRunner | `"false"` | deploy, ssh | `"true"` or `"1"` to prefer private IP |
| `BUILD_ON_TARGET` | mkRunner | `"true"` | deploy | `"false"` or `"0"` to build locally |

### User-Provided Variables

These are NOT set by mkRunner and must be provided by the user:

| Variable | Description |
|----------|-------------|
| `AWS_ACCESS_KEY_ID` | AWS credentials (alternative to `terraform.tfvars`) |
| `AWS_SECRET_ACCESS_KEY` | AWS credentials (alternative to `terraform.tfvars`) |
| `TF_VAR_*` | Terraform variables passed via environment |

## AWS Credential Setup

Inframan passes the full environment to Terraform, so any standard credential method works:

### Option 1: `terraform.tfvars` (Recommended)

Create a `terraform.tfvars` file in the project's Terraform directory:

```bash
mkdir -p .inframan/<project-name>/terraform
cat > .inframan/<project-name>/terraform/terraform.tfvars <<EOF
aws_access_key = "AKIA..."
aws_secret_key = "..."
ssh_public_key = "ssh-ed25519 AAAA..."
EOF
```

### Option 2: Environment Variables

```bash
export TF_VAR_aws_access_key="AKIA..."
export TF_VAR_aws_secret_key="..."
export TF_VAR_ssh_public_key="ssh-ed25519 AAAA..."
nix run .#prod -- infra
```

## IP Resolution Logic

When inframan needs a target IP (for `deploy` or `ssh`), it resolves it from Terraform output:

1. Parse `terraform output -json`
2. If the output contains an `instances` map, use it (multi-instance mode)
3. If `USE_PRIVATE_IP=true` **or** `SSH_PROXY_JUMP` is configured:
   - Use `private_ip` if available
4. Otherwise, use `public_ip`
5. Error if no IP is found

This means setting `sshProxyJump` in `mkRunner` automatically switches to private IPs -- you don't need to also set `usePrivateIp` when using a bastion.

## Directory Structure

```
.inframan/
└── <project-name>/           # PROJECT_NAME (default: "default")
    ├── terraform/
    │   ├── config.tf.json    # Copied from Terranix output
    │   ├── .terraform/       # Terraform plugin cache (after init)
    │   ├── terraform.tfstate # Local state (if not using remote backend)
    │   └── terraform.tfvars  # User-provided credentials (optional)
    └── colmena/
        └── hive.nix          # Generated on each deploy
```

## Defaults Summary

| Setting | Default | Override |
|---------|---------|---------|
| Project name | `"default"` | `projectName` in mkRunner |
| Target architecture | `"x86_64-linux"` | `targetSystem` in mkRunner |
| Build on target | `true` | `buildOnTarget` in mkRunner |
| Use private IP | `false` | `usePrivateIp` in mkRunner or proxy jump configured |
| SSH user (ssh command) | `"root"` | `--user` flag |
