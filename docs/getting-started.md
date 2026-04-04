# Getting Started

This guide walks you through deploying a NixOS instance on AWS using inframan.

## Prerequisites

- **Nix 2.4+** with flakes enabled (`experimental-features = nix-command flakes` in `~/.config/nix/nix.conf`)
- **AWS account** with access key and secret key
- **SSH key pair** for instance access

## Project Setup

Create a new directory for your project:

```bash
mkdir my-infra && cd my-infra
git init
```

### 1. Create `flake.nix`

This is the entry point. It uses `inframan.lib.mkRunner` to create a project-specific runner that bundles your infrastructure and machine configs together.

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    inframan.url = "github:iivel-inc/inframan";
  };

  outputs = { self, nixpkgs, inframan, ... }: {
    apps.x86_64-linux.default = {
      type = "app";
      program = "${inframan.lib.mkRunner {
        system = "x86_64-linux";
        infraConfig = ./infrastructure.nix;
        machineConfig = ./machine.nix;
      }}/bin/runner";
    };
  };
}
```

### 2. Create `infrastructure.nix`

This Terranix file defines your AWS infrastructure. It compiles to Terraform JSON at Nix evaluation time.

```nix
{
  provider.aws = {
    region = "us-east-1";
    access_key = "\${var.aws_access_key}";
    secret_key = "\${var.aws_secret_key}";
  };

  variable.aws_access_key = {
    type = "string";
    sensitive = true;
  };

  variable.aws_secret_key = {
    type = "string";
    sensitive = true;
  };

  variable.ssh_public_key = {
    type = "string";
  };

  # Latest NixOS AMI
  data.aws_ami.nixos = {
    most_recent = true;
    owners = [ "427812963091" ];
    filter = [
      { name = "name"; values = [ "nixos/24.05*" ]; }
      { name = "architecture"; values = [ "x86_64" ]; }
    ];
  };

  resource.aws_key_pair.deployer = {
    key_name = "inframan-deployer";
    public_key = "\${var.ssh_public_key}";
  };

  resource.aws_security_group.main = {
    name = "inframan-sg";
    ingress = [
      { from_port = 22; to_port = 22; protocol = "tcp"; cidr_blocks = [ "0.0.0.0/0" ]; ipv6_cidr_blocks = []; prefix_list_ids = []; security_groups = []; self = false; description = "SSH"; }
      { from_port = 80; to_port = 80; protocol = "tcp"; cidr_blocks = [ "0.0.0.0/0" ]; ipv6_cidr_blocks = []; prefix_list_ids = []; security_groups = []; self = false; description = "HTTP"; }
    ];
    egress = [
      { from_port = 0; to_port = 0; protocol = "-1"; cidr_blocks = [ "0.0.0.0/0" ]; ipv6_cidr_blocks = []; prefix_list_ids = []; security_groups = []; self = false; description = "All outbound"; }
    ];
  };

  resource.aws_instance.main = {
    ami = "\${data.aws_ami.nixos.id}";
    instance_type = "t3.micro";
    key_name = "\${aws_key_pair.deployer.key_name}";
    vpc_security_group_ids = [ "\${aws_security_group.main.id}" ];
    root_block_device = { volume_size = 20; volume_type = "gp3"; };
    tags = { Name = "inframan-instance"; ManagedBy = "inframan"; };
  };

  # Required output -- inframan reads this to know where to deploy
  output.public_ip = {
    value = "\${aws_instance.main.public_ip}";
  };
}
```

### 3. Create `machine.nix`

This NixOS module defines what gets deployed to the instance via Colmena.

```nix
{ config, pkgs, ... }:

{
  system.stateVersion = "24.05";

  boot.loader.grub.device = "nodev";
  boot.loader.grub.efiSupport = true;
  boot.loader.efi.canTouchEfiVariables = true;

  networking.hostName = "my-server";
  networking.firewall.allowedTCPPorts = [ 22 80 ];

  services.openssh = {
    enable = true;
    settings.PermitRootLogin = "prohibit-password";
  };

  users.users.root.openssh.authorizedKeys.keys = [
    "ssh-ed25519 AAAA... your-public-key-here"
  ];

  services.nginx = {
    enable = true;
    virtualHosts."default" = {
      default = true;
      root = pkgs.writeTextDir "index.html" "<h1>Hello from Inframan!</h1>";
    };
  };

  environment.systemPackages = with pkgs; [ vim htop curl ];

  nix.settings.experimental-features = [ "nix-command" "flakes" ];
}
```

## Deploy

### Step 1: Set up credentials

Create a `terraform.tfvars` file in the terraform workspace directory:

```bash
mkdir -p .inframan/default/terraform
cat > .inframan/default/terraform/terraform.tfvars <<EOF
aws_access_key = "YOUR_ACCESS_KEY"
aws_secret_key = "YOUR_SECRET_KEY"
ssh_public_key = "ssh-ed25519 AAAA... your-public-key"
EOF
```

### Step 2: Provision infrastructure

```bash
nix run . -- infra
```

This copies the Terranix JSON to `.inframan/default/terraform/config.tf.json`, runs `terraform init`, then `terraform apply`. Terraform will show a plan and ask for confirmation.

### Step 3: Deploy NixOS configuration

```bash
nix run . -- deploy
```

This fetches the instance IP from Terraform state, generates an ephemeral `hive.nix`, and runs `colmena apply` to deploy your NixOS configuration.

### Step 4: Verify

```bash
# SSH into the instance
nix run . -- ssh default

# Or curl the web server
curl http://$(cd .inframan/default/terraform && terraform output -raw public_ip)
```

## Cleanup

Destroy all provisioned infrastructure:

```bash
nix run . -- destroy
```

## Next Steps

- [Architecture](architecture.md) -- Understand how the pieces fit together
- [Configuration](configuration.md) -- All available options
- [Multi-Project Setup](multi-project.md) -- Managing multiple environments
- [SSH & Networking](ssh-and-networking.md) -- Advanced SSH and bastion host setups
