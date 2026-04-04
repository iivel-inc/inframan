# Inframan Documentation

Inframan is a Go CLI tool that bridges [Terranix](https://terranix.org/) (Infrastructure as Nix) and [Colmena](https://colmena.cli.rs/) (NixOS deployment), enabling declarative infrastructure provisioning and NixOS configuration management in a unified workflow. Define your infrastructure and machine configuration purely in Nix, then provision and deploy with simple CLI commands.

## Documentation

### Getting Started

- [Getting Started](getting-started.md) -- Quickstart guide from zero to a deployed NixOS instance

### Core Concepts

- [Architecture](architecture.md) -- How inframan works: layers, data flow, and project structure
- [Configuration](configuration.md) -- All configuration options, environment variables, and defaults
- [Nix Integration](nix-integration.md) -- Deep dive into `flake.nix`, `lib.mkRunner`, and the Nix ecosystem

### Usage Guides

- [Command Reference](commands.md) -- Every CLI command with flags, examples, and behavior notes
- [Infrastructure Templates](infrastructure-templates.md) -- Writing Terranix configs for AWS and other providers
- [Machine Configuration](machine-configuration.md) -- Writing NixOS modules deployed via Colmena
- [Multi-Project Setup](multi-project.md) -- Managing multiple projects and AWS accounts
- [SSH & Networking](ssh-and-networking.md) -- SSH access, bastion hosts, private networking

### Development

- [Development Guide](development.md) -- Contributing, adding commands, code conventions
- [Troubleshooting](troubleshooting.md) -- Common errors and how to fix them

## Suggested Reading Order

**New to inframan?** Start with [Getting Started](getting-started.md), then read [Architecture](architecture.md) and [Configuration](configuration.md).

**Setting up a project?** Read [Infrastructure Templates](infrastructure-templates.md), [Machine Configuration](machine-configuration.md), and [Nix Integration](nix-integration.md).

**Contributing to inframan?** Read [Architecture](architecture.md) and [Development Guide](development.md).
