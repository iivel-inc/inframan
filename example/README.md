# Inframan Example

This example demonstrates how to use inframan to provision AWS infrastructure with Terranix/Terraform and deploy a NixOS configuration with Colmena.

## Files

- `flake.nix` - Nix flake that uses `inframan.lib.mkRunner`
- `infrastructure.nix` - Terranix configuration defining AWS resources
- `machine.nix` - NixOS module to deploy on the provisioned instance

## Prerequisites

1. **AWS credentials** - Set `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables
2. **SSH key** - Have an SSH public key ready for instance access
3. **Nix with flakes** - Nix 2.4+ with experimental features enabled

## Usage

### 1. Set up environment

```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"

# Create a terraform.tfvars file for the SSH key
echo 'ssh_public_key = "ssh-ed25519 AAAA... your-key"' > terraform.tfvars
```

### 2. Edit machine.nix

Add your SSH public key to `users.users.root.openssh.authorizedKeys.keys` in `machine.nix`.

### 3. Provision infrastructure

```bash
# From the example directory
nix run . -- infra
```

This will:
- Compile `infrastructure.nix` to `config.tf.json` via Terranix
- Run `terraform init` and `terraform apply`
- Provision the AWS EC2 instance

### 4. Deploy NixOS configuration

```bash
nix run . -- deploy
```

This will:
- Read the instance IP from `terraform output`
- Generate an ephemeral `hive.nix` with the target IP
- Run `colmena apply` to deploy the NixOS configuration

### 5. Verify

After deployment, visit `http://<instance-ip>` to see the nginx welcome page.

## Customization

### Different instance type

Edit `infrastructure.nix`:
```nix
variable.instance_type.default = "t3.small";
```

### Additional services

Add services to `machine.nix`:
```nix
services.postgresql.enable = true;
```

### Multiple instances

Extend `infrastructure.nix` with multiple `aws_instance` resources and update `machine.nix` to configure them differently based on hostname.

### Multiple projects

To manage multiple environments (e.g., staging and production), add the `projectName` parameter to `mkRunner` in `flake.nix`:

```nix
packages.${system}.staging = inframan.lib.mkRunner {
  inherit system;
  infraConfig = ./infrastructure.nix;
  machineConfig = ./machine.nix;
  projectName = "staging";
};

packages.${system}.production = inframan.lib.mkRunner {
  inherit system;
  infraConfig = ./infrastructure-prod.nix;
  machineConfig = ./machine-prod.nix;
  projectName = "production";
};
```

Each project will have its own isolated directory:
- `.inframan/staging/terraform/` - Staging Terraform state
- `.inframan/production/terraform/` - Production Terraform state

## Cleanup

To destroy the infrastructure:

```bash
# For default project
cd .inframan/default/terraform
terraform destroy

# For named project (e.g., "staging")
cd .inframan/staging/terraform
terraform destroy
```

