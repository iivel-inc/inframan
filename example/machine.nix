# NixOS configuration module for the target machine
# This is deployed via Colmena after infrastructure is provisioned
{ config, pkgs, lib, ... }:

{
  # System basics
  system.stateVersion = "24.05";

  # Boot configuration for AWS
  boot.loader.grub.device = "nodev";
  boot.loader.grub.efiSupport = true;
  boot.loader.efi.canTouchEfiVariables = true;

  # Networking
  networking.hostName = "inframan-node";
  networking.firewall = {
    enable = true;
    allowedTCPPorts = [ 22 80 443 ];
  };

  # SSH access
  services.openssh = {
    enable = true;
    settings = {
      PermitRootLogin = "prohibit-password";
      PasswordAuthentication = false;
    };
  };

  # Root user SSH key (add your public key here)
  users.users.root.openssh.authorizedKeys.keys = [
    # "ssh-ed25519 AAAA... your-key-here"
  ];

  # Example: Deploy a simple web server
  services.nginx = {
    enable = true;
    virtualHosts."default" = {
      default = true;
      root = pkgs.writeTextDir "index.html" ''
        <!DOCTYPE html>
        <html>
          <head><title>Inframan Deployed!</title></head>
          <body>
            <h1>Hello from Inframan!</h1>
            <p>This NixOS server was deployed using:</p>
            <ul>
              <li>Terranix - Infrastructure as Nix</li>
              <li>OpenTofu - Infrastructure provisioning</li>
              <li>Colmena - NixOS deployment</li>
              <li>Inframan - The Go bridge connecting them</li>
            </ul>
          </body>
        </html>
      '';
    };
  };

  # System packages
  environment.systemPackages = with pkgs; [
    vim
    htop
    git
    curl
    wget
  ];

  # Automatic garbage collection
  nix.gc = {
    automatic = true;
    dates = "weekly";
    options = "--delete-older-than 30d";
  };

  # Enable flakes
  nix.settings.experimental-features = [ "nix-command" "flakes" ];
}

