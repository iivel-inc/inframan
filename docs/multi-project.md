# Multi-Project Setup

Inframan supports managing multiple independent projects in the same workspace. Each project has its own infrastructure config, machine config, Terraform state, and Colmena configuration.

## Project Isolation

Each project gets a fully isolated directory under `.inframan/`:

```
.inframan/
├── production/
│   ├── terraform/          # Independent Terraform state
│   │   ├── config.tf.json
│   │   ├── .terraform/
│   │   ├── terraform.tfstate
│   │   └── terraform.tfvars
│   └── colmena/
│       └── hive.nix
├── staging/
│   ├── terraform/
│   └── colmena/
└── development/
    ├── terraform/
    └── colmena/
```

Projects never share state. Running `inframan destroy` for one project does not affect others.

## Defining Multiple Runners

In your `flake.nix`, create a separate `lib.mkRunner` for each project:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    inframan.url = "github:iivel-inc/inframan";
  };

  outputs = { self, nixpkgs, inframan, ... }:
    let
      system = "x86_64-linux";
    in
    {
      # Production
      packages.${system}.prod = inframan.lib.mkRunner {
        inherit system;
        infraConfig = ./infra-prod.nix;
        machineConfig = ./machine-prod.nix;
        projectName = "production";
        sshKeyPath = "./keys/prod.pem";
      };

      apps.${system}.prod = {
        type = "app";
        program = "${self.packages.${system}.prod}/bin/runner";
      };

      # Staging
      packages.${system}.staging = inframan.lib.mkRunner {
        inherit system;
        infraConfig = ./infra-staging.nix;
        machineConfig = ./machine-staging.nix;
        projectName = "staging";
      };

      apps.${system}.staging = {
        type = "app";
        program = "${self.packages.${system}.staging}/bin/runner";
      };

      # Default points to production
      apps.${system}.default = self.apps.${system}.prod;
    };
}
```

## Running Specific Projects

```bash
# Provision production
nix run .#prod -- infra

# Deploy to staging
nix run .#staging -- deploy

# SSH to production
nix run .#prod -- ssh production

# List all instances across all projects
nix run .#prod -- ssh --list

# Destroy staging
nix run .#staging -- destroy
```

## Per-Project Credentials

Each project can have its own credentials via `terraform.tfvars`:

```bash
# Production credentials
cat > .inframan/production/terraform/terraform.tfvars <<EOF
aws_access_key = "AKIA...prod..."
aws_secret_key = "..."
ssh_public_key = "ssh-ed25519 AAAA..."
EOF

# Staging credentials (different AWS account)
cat > .inframan/staging/terraform/terraform.tfvars <<EOF
aws_access_key = "AKIA...staging..."
aws_secret_key = "..."
ssh_public_key = "ssh-ed25519 AAAA..."
EOF
```

## Multi-Account AWS

A common pattern is one project per AWS account. Each project has:
- Its own `infrastructure-<name>.nix` with account-specific provider config and resources
- Its own `machine-<name>.nix` with environment-specific NixOS config
- Its own AWS credentials in `terraform.tfvars`

See [example/](../example/) for a complete multi-account setup with two AWS accounts, and [example/MULTI_ACCOUNT_GUIDE.md](../example/MULTI_ACCOUNT_GUIDE.md) for a detailed walkthrough.

## Adding a New Project

1. Create `infrastructure-<name>.nix` (copy and modify from an existing one)
2. Create `machine-<name>.nix` (copy and modify from an existing one)
3. Add a new `lib.mkRunner` block in `flake.nix` with a unique `projectName`
4. Set up credentials: `mkdir -p .inframan/<name>/terraform && vim .inframan/<name>/terraform/terraform.tfvars`
5. Run: `nix run .#<name> -- infra && nix run .#<name> -- deploy`

## SSH Across Projects

The `inframan ssh` command works across all projects:

```bash
# List all instances from all projects
inframan ssh --list
# Output:
#   production         54.123.45.67
#   staging            34.56.78.90

# Connect to a specific project
inframan ssh production
inframan ssh staging

# Multi-instance projects use project/instance syntax
inframan ssh production/web-1
inframan ssh production/db-1
```
