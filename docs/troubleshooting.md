# Troubleshooting

Common errors and how to fix them.

## "INFRA_CONFIG_JSON environment variable is not set"

**Cause:** Running the `inframan` binary directly instead of through `nix run`.

**Fix:** Use the Nix runner which sets all required env vars:

```bash
nix run .#<project> -- <command>
```

Or set the variable manually for development:

```bash
export INFRA_CONFIG_JSON=/path/to/config.tf.json
```

## "NIXOS_MODULE_PATH environment variable is not set"

**Cause:** Same as above -- running `inframan deploy` directly.

**Fix:** Use `nix run`, or set manually:

```bash
export NIXOS_MODULE_PATH=/path/to/machine.nix
```

## "terraform init failed"

**Possible causes:**
- **Network issues** -- Terraform needs to download providers on first init
- **Backend configuration errors** -- Check S3 bucket exists and credentials are correct
- **Missing credentials** -- Ensure `terraform.tfvars` or `TF_VAR_*` environment variables are set

**Debug:**

```bash
cd .inframan/<project>/terraform
terraform init  # Run manually to see full error output
```

## "No IP found in terraform output"

**Cause:** Infrastructure hasn't been provisioned, or the `infrastructure.nix` doesn't include the required outputs.

**Fix:**
1. Run `inframan infra` first to provision infrastructure
2. Ensure your `infrastructure.nix` includes at least one of:
   - `output.public_ip` (single instance)
   - `output.private_ip` (for private networking)
   - `output.instances` (multi-instance)

**Debug:**

```bash
cd .inframan/<project>/terraform
terraform output -json
```

## "colmena apply failed"

**Possible causes:**

### SSH connection failure
- **Key not authorized** -- Ensure `users.users.root.openssh.authorizedKeys.keys` in your `machine.nix` includes your SSH public key
- **Security group blocking SSH** -- Ensure port 22 is open in your security group
- **Wrong IP** -- If using private IP mode, ensure you have network access (VPN, bastion)

### Build failure
- **Target out of disk space** -- If `buildOnTarget = true`, the target needs enough disk for compilation
- **Invalid NixOS module** -- Check your `machine.nix` for syntax errors
- **Architecture mismatch** -- Ensure `targetSystem` matches the target machine

**Debug:**

```bash
# Check the generated hive
cat .inframan/<project>/colmena/hive.nix

# Test SSH manually
ssh -i <key> root@<ip>

# Validate the hive
cd .inframan/<project>/colmena
colmena eval -f hive.nix -E '{ nodes, ... }: nodes'
```

## "project does not exist"

**Cause:** Typo in project name, or the project hasn't been initialized.

**Fix:**

```bash
# Check what projects exist
ls .inframan/

# Initialize the project
nix run .#<project> -- init
```

## SSH connection timeout

**Possible causes:**
- Security group doesn't allow SSH (port 22) from your IP
- Using public IP when instance is in a private subnet
- Bastion host is unreachable
- Instance hasn't finished booting

**Fix:**

```bash
# Check which IP inframan is using
cd .inframan/<project>/terraform
terraform output -json | jq '.public_ip.value, .private_ip.value'

# Test connectivity
nc -zv <ip> 22

# If using bastion, test bastion first
ssh admin@bastion.example.com
```

## "No instances found"

**Cause:** Haven't run `inframan infra` yet, or Terraform state is empty.

**Fix:**

```bash
nix run .#<project> -- infra   # Provision infrastructure first
nix run .#<project> -- ssh --list  # Then list instances
```

## Nix evaluation errors

**Cause:** Syntax errors in `infrastructure.nix`, `machine.nix`, or `flake.nix`.

**Debug:**

```bash
# Check flake evaluation
nix flake show

# Check specific output
nix eval .#apps.x86_64-linux.default

# For Terranix config issues
nix build .#<project> --print-build-logs
```

## "vendorHash mismatch" when building

**Cause:** Go dependencies changed but `vendorHash` in `flake.nix` wasn't updated.

**Fix:**

```bash
nix build 2>&1 | grep "got:"
# Copy the sha256-... hash from the output
# Update vendorHash in flake.nix
```

## Inspecting State

When debugging, these commands help understand the current state:

```bash
# View what's in the terraform workspace
ls -la .inframan/<project>/terraform/

# View terraform state
cd .inframan/<project>/terraform && terraform show

# View all outputs
cd .inframan/<project>/terraform && terraform output -json

# View the generated Colmena hive
cat .inframan/<project>/colmena/hive.nix

# View the Terranix-generated config
cat .inframan/<project>/terraform/config.tf.json | jq .
```
