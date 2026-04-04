---
name: write-nix-config
description: Guide for writing infrastructure.nix (Terranix) and machine.nix (NixOS) files for inframan
trigger: When the user asks to create, modify, or write Nix configuration files for inframan
---

# Writing Nix Configuration for Inframan

Inframan uses two Nix files per project: `infrastructure.nix` (Terranix, compiles to Terraform JSON) and `machine.nix` (NixOS module, deployed via Colmena).

## infrastructure.nix (Terranix)

### Required Structure

```nix
{
  # Provider (at minimum one cloud provider)
  provider.aws = {
    region = "us-east-1";
    access_key = "\${var.aws_access_key}";
    secret_key = "\${var.aws_secret_key}";
  };

  # Variables for credentials (passed via terraform.tfvars)
  variable.aws_access_key = { type = "string"; sensitive = true; };
  variable.aws_secret_key = { type = "string"; sensitive = true; };

  # Resources...

  # REQUIRED: At least one of these outputs for inframan to discover IPs
  output.public_ip = { value = "\${aws_instance.main.public_ip}"; };
}
```

### Output Formats

**Single instance** (legacy):

```nix
output.public_ip = { value = "\${aws_instance.main.public_ip}"; };
output.private_ip = { value = "\${aws_instance.main.private_ip}"; };  # optional
```

**Multiple instances**:

```nix
output.instances = {
  value = {
    "web-1" = "\${aws_instance.web[0].public_ip}";
    "db-1" = "\${aws_instance.db.public_ip}";
  };
};
```

### Variable Interpolation

Escape `$` for Terraform references in Nix strings:

```nix
# Correct
ami = "\${data.aws_ami.nixos.id}";

# Wrong (Nix will try to evaluate this)
ami = "${data.aws_ami.nixos.id}";
```

### NixOS AMI Data Source

```nix
data.aws_ami.nixos = {
  most_recent = true;
  owners = [ "427812963091" ];  # Official NixOS AMI owner
  filter = [
    { name = "name"; values = [ "nixos/24.05*" ]; }
    { name = "architecture"; values = [ "x86_64" ]; }
  ];
};
```

### Security Group (all fields required to avoid diff issues)

```nix
resource.aws_security_group.main = {
  name = "my-sg";
  ingress = [{
    description = "SSH"; from_port = 22; to_port = 22; protocol = "tcp";
    cidr_blocks = [ "0.0.0.0/0" ];
    ipv6_cidr_blocks = []; prefix_list_ids = []; security_groups = []; self = false;
  }];
  egress = [{
    description = "All"; from_port = 0; to_port = 0; protocol = "-1";
    cidr_blocks = [ "0.0.0.0/0" ];
    ipv6_cidr_blocks = []; prefix_list_ids = []; security_groups = []; self = false;
  }];
};
```

### Remote Backend (optional)

```nix
terraform.backend.s3 = {
  bucket = "my-state-bucket";
  key = "inframan/<project>/terraform.tfstate";
  region = "us-east-1";
  encrypt = true;
};
```

## machine.nix (NixOS Module)

### Required Structure

```nix
{ config, pkgs, lib, ... }:

{
  system.stateVersion = "24.05";  # Match your NixOS AMI version

  # Boot (for AWS EC2)
  boot.loader.grub.device = "nodev";
  boot.loader.grub.efiSupport = true;
  boot.loader.efi.canTouchEfiVariables = true;

  # SSH (required for Colmena deployment)
  services.openssh.enable = true;
  services.openssh.settings.PermitRootLogin = "prohibit-password";

  # Root SSH key (required for Colmena)
  users.users.root.openssh.authorizedKeys.keys = [
    "ssh-ed25519 AAAA... key-here"
  ];

  # Firewall
  networking.firewall.allowedTCPPorts = [ 22 ];

  # Flakes
  nix.settings.experimental-features = [ "nix-command" "flakes" ];
}
```

### Do NOT Set These

Inframan injects these via the generated `hive.nix`:
- `deployment.targetHost`
- `deployment.targetUser`
- `deployment.buildOnTarget`
- `deployment.sshOptions`

## flake.nix Integration

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    inframan.url = "github:iivel-inc/inframan";
  };

  outputs = { self, nixpkgs, inframan, ... }:
    let system = "x86_64-linux";
    in {
      apps.${system}.default = {
        type = "app";
        program = "${inframan.lib.mkRunner {
          inherit system;
          infraConfig = ./infrastructure.nix;
          machineConfig = ./machine.nix;
          # Optional:
          # projectName = "production";
          # sshKeyPath = "./keys/deploy.pem";
          # sshProxyJump = "admin@bastion:22";
          # buildOnTarget = true;
          # usePrivateIp = false;
          # targetSystem = "x86_64-linux";
        }}/bin/runner";
      };
    };
}
```

## Reference

- Complete infrastructure example: `example/infrastructure.nix` (also `example/infrastructure-account1.nix`, `example/infrastructure-account2.nix`)
- Complete machine example: `example/machine.nix` (also `example/machine-account1.nix`, `example/machine-account2.nix`)
- Multi-project flake example: `example/flake.nix`
- Full documentation: `docs/infrastructure-templates.md`, `docs/machine-configuration.md`
