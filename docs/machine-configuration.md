# Writing NixOS Machine Configs

Machine configuration in inframan is a standard NixOS module deployed to the target instance via Colmena. The module defines everything about the running system: services, packages, users, networking, and more.

## Module Format

A machine config is a NixOS module -- a function that takes module arguments and returns an attribute set:

```nix
{ config, pkgs, lib, ... }:

{
  # NixOS configuration here
}
```

## What Inframan Injects

When you run `inframan deploy`, it generates an ephemeral `hive.nix` that wraps your module with deployment-specific settings. You do NOT need to set these in your machine config:

```nix
# These are injected by inframan -- don't set them yourself:
deployment.targetHost = "1.2.3.4";       # From terraform output
deployment.targetUser = "root";
deployment.buildOnTarget = true;          # From BUILD_ON_TARGET
deployment.sshOptions = [ "-i" "..." ];  # From SSH_* env vars
```

The generated `hive.nix` imports your module and adds these deployment attributes. The `meta.nixpkgs` is set to use the `TARGET_SYSTEM` architecture.

## Recommended Base Config

Every machine config should include:

```nix
{ config, pkgs, ... }:

{
  # Pin the NixOS state version to match your AMI
  system.stateVersion = "24.05";

  # Boot loader (for AWS EC2 with NixOS AMI)
  boot.loader.grub.device = "nodev";
  boot.loader.grub.efiSupport = true;
  boot.loader.efi.canTouchEfiVariables = true;

  # Networking
  networking.hostName = "my-server";
  networking.firewall = {
    enable = true;
    allowedTCPPorts = [ 22 ];  # At minimum, keep SSH open
  };

  # SSH access (required for deployment and management)
  services.openssh = {
    enable = true;
    settings = {
      PermitRootLogin = "prohibit-password";
      PasswordAuthentication = false;
    };
  };

  # Root SSH authorized keys (required for Colmena deployment)
  users.users.root.openssh.authorizedKeys.keys = [
    "ssh-ed25519 AAAA... your-public-key-here"
  ];

  # Enable flakes on the target
  nix.settings.experimental-features = [ "nix-command" "flakes" ];
}
```

## Adding Services

Add any NixOS service using the standard module system:

### Web Server (nginx)

```nix
services.nginx = {
  enable = true;
  virtualHosts."example.com" = {
    root = "/var/www/example.com";
    locations."/" = {
      tryFiles = "$uri $uri/ =404";
    };
  };
};

networking.firewall.allowedTCPPorts = [ 80 443 ];
```

### Database (PostgreSQL)

```nix
services.postgresql = {
  enable = true;
  package = pkgs.postgresql_16;
  ensureDatabases = [ "myapp" ];
  ensureUsers = [
    { name = "myapp"; ensureDBOwnership = true; }
  ];
};
```

### Monitoring

```nix
services.prometheus = {
  enable = true;
  port = 9090;
};

services.grafana = {
  enable = true;
  settings.server.http_port = 3000;
};
```

## System Packages

```nix
environment.systemPackages = with pkgs; [
  vim
  htop
  git
  curl
  wget
  tmux
  jq
];
```

## Automatic Maintenance

```nix
# Garbage collection
nix.gc = {
  automatic = true;
  dates = "weekly";
  options = "--delete-older-than 30d";
};

# Automatic NixOS upgrades (optional)
system.autoUpgrade = {
  enable = true;
  allowReboot = false;
};
```

## Additional Users

```nix
users.users.deploy = {
  isNormalUser = true;
  extraGroups = [ "wheel" ];
  openssh.authorizedKeys.keys = [
    "ssh-ed25519 AAAA... deploy-key"
  ];
};
```

## Build Modes

The `buildOnTarget` parameter (default: `true`) controls where NixOS is built:

- **`buildOnTarget = true`** (default): Colmena copies the Nix expressions to the target and builds there. Requires the target to have enough resources for compilation.
- **`buildOnTarget = false`**: Colmena builds locally and copies the closure to the target via `nix-copy-closure`. Requires the local machine to match the target architecture (or use cross-compilation).

Set this in `lib.mkRunner`:

```nix
inframan.lib.mkRunner {
  # ...
  buildOnTarget = false;  # Build locally, copy to target
}
```

## Complete Example

See [example/machine.nix](../example/machine.nix) for a complete working example with boot config, networking, SSH, nginx, system packages, and garbage collection.
