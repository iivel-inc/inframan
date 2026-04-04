# SSH & Networking

Inframan provides SSH access to provisioned instances with support for bastion hosts, private networking, and multi-instance projects.

## Basic SSH

```bash
# Connect to a single-instance project
inframan ssh my-project

# Connect as a specific user (default: root)
inframan ssh my-project --user nixos

# Use a specific identity file
inframan ssh my-project --identity ~/.ssh/id_ed25519
```

## Multi-Instance SSH

For projects that output an `instances` map (see [Infrastructure Templates](infrastructure-templates.md)), use `project/instance` syntax:

```bash
# List all instances
inframan ssh --list
# Output:
#   production/web-1   54.123.45.67
#   production/db-1    10.0.1.50
#   staging            34.56.78.90

# Connect to a specific instance
inframan ssh production/web-1
inframan ssh production/db-1
```

If a project has multiple instances and you don't specify which one, inframan will error and list available instances.

## Remote Commands

Run commands on the remote host without an interactive session:

```bash
# Run a single command
inframan ssh my-project -- hostname
inframan ssh my-project -- cat /etc/os-release

# Pipe input to a remote command
echo "hello" | inframan ssh my-project -- cat

# Run multi-word commands
inframan ssh my-project -- systemctl status nginx
```

## SSH Configuration Priority

When building SSH arguments, inframan checks in this order:

1. **`SSH_CONFIG_PATH`** -- If set, uses `-F <config>` and skips all other SSH options (no `StrictHostKeyChecking`, no `UserKnownHostsFile`, no `LogLevel`). Your config file has full control.
2. **`--identity` flag** -- If provided on the command line, uses `-i <file>`
3. **`SSH_KEY_PATH`** -- If set in environment, uses `-i <file>`
4. **Default** -- No key specified, SSH uses its default behavior

When `SSH_CONFIG_PATH` is NOT set, these convenience options are added automatically:
- `-o StrictHostKeyChecking=accept-new` -- Auto-accept new host keys
- `-o UserKnownHostsFile=/dev/null` -- Don't pollute known_hosts
- `-o LogLevel=ERROR` -- Suppress warnings

## Proxy Jump / Bastion Host

For instances behind a bastion/jump host, configure `sshProxyJump` in `lib.mkRunner`:

```nix
inframan.lib.mkRunner {
  # ...
  sshProxyJump = "admin@bastion.example.com:22";
}
```

This adds `-J admin@bastion.example.com:22` to SSH commands and **automatically switches to private IPs** from Terraform output. You don't need to also set `usePrivateIp`.

The proxy jump is used by:
- `inframan ssh` -- For interactive SSH sessions
- `inframan deploy` -- Both in the generated `hive.nix` (`deployment.sshOptions`) and `NIX_SSHOPTS` for `nix-copy-closure`

## Private IP Mode

For instances reachable via private network (VPN, VPC peering) without a bastion, use `usePrivateIp`:

```nix
inframan.lib.mkRunner {
  # ...
  usePrivateIp = true;
}
```

This makes inframan prefer `private_ip` from Terraform output instead of `public_ip`. Your `infrastructure.nix` must output `private_ip`:

```nix
output.private_ip = {
  value = "\${aws_instance.main.private_ip}";
};
```

## IP Resolution Summary

| Configuration | IP Used |
|--------------|---------|
| Default | `public_ip` |
| `usePrivateIp = true` | `private_ip` (falls back to `public_ip` if not available) |
| `sshProxyJump` set | `private_ip` (target is behind bastion) |
| Both set | `private_ip` |

## Colmena SSH Integration

During `inframan deploy`, SSH options flow to Colmena in two ways:

1. **`hive.nix`** -- The generated hive includes `deployment.sshOptions` with all configured SSH flags
2. **`NIX_SSHOPTS`** -- Environment variable set for `nix-copy-closure` (used when `buildOnTarget = false` to copy the closure to the target)

Both paths include the same SSH configuration: config file, identity file, proxy jump, and `StrictHostKeyChecking=accept-new`.

## Process Replacement

The `inframan ssh` command uses `syscall.Exec()` to replace the inframan process with the SSH binary. This is not `os/exec` (which creates a child process) -- it completely replaces the current process. Benefits:

- Full terminal control (TTY allocation, colors, resize)
- Signal handling works correctly (Ctrl+C goes to SSH, not inframan)
- Piped input/output works as expected
- No process overhead from a parent wrapper

## Network Architecture Examples

### Direct Public Access

```
Your Machine ──(SSH)──▶ Public IP ──▶ Instance
```

```nix
inframan.lib.mkRunner {
  # Default: uses public_ip, no proxy
  system = "x86_64-linux";
  infraConfig = ./infrastructure.nix;
  machineConfig = ./machine.nix;
}
```

### Bastion / Jump Host

```
Your Machine ──(SSH)──▶ Bastion (public) ──(SSH)──▶ Instance (private)
```

```nix
inframan.lib.mkRunner {
  system = "x86_64-linux";
  infraConfig = ./infrastructure.nix;
  machineConfig = ./machine.nix;
  sshProxyJump = "admin@bastion.example.com:22";
  # usePrivateIp is automatic when sshProxyJump is set
}
```

### VPN / VPC Peering

```
Your Machine ──(VPN)──▶ Private Network ──▶ Instance (private)
```

```nix
inframan.lib.mkRunner {
  system = "x86_64-linux";
  infraConfig = ./infrastructure.nix;
  machineConfig = ./machine.nix;
  usePrivateIp = true;
}
```
