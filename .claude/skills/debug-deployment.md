---
name: debug-deployment
description: Diagnostic checklist for debugging failed inframan deployments
trigger: When the user reports a deployment failure, SSH issue, or infrastructure error
---

# Debugging Inframan Deployments

## Diagnostic Checklist

### 1. Check if infrastructure is provisioned

```bash
ls .inframan/<project>/terraform/
# Should contain: config.tf.json, .terraform/, terraform.tfstate (or remote backend)
```

If missing, the project hasn't been initialized or provisioned:

```bash
nix run .#<project> -- init   # Just initialize
nix run .#<project> -- infra  # Provision infrastructure
```

### 2. Check Terraform state

```bash
cd .inframan/<project>/terraform
terraform output -json
```

Expected output should include at least one of:
- `public_ip` with a non-empty value
- `private_ip` with a non-empty value
- `instances` map with IP addresses

If empty or missing, infrastructure provisioning failed or outputs are missing from `infrastructure.nix`.

### 3. Verify NixOS module

```bash
# Check the NIXOS_MODULE_PATH points to a valid file
ls -la $(nix eval --raw .#<project>.program 2>/dev/null | xargs -I{} sh -c 'grep NIXOS_MODULE_PATH {}')
# Or just check the file referenced in your flake.nix machineConfig
```

### 4. For SSH failures

```bash
# Test connectivity to the IP
nc -zv <ip> 22

# Test SSH manually
ssh -v -i <key-path> root@<ip>

# Check if using the right IP (public vs private)
cd .inframan/<project>/terraform
terraform output -json | jq '{public_ip: .public_ip.value, private_ip: .private_ip.value}'
```

Common SSH issues:
- **Key not authorized**: Check `users.users.root.openssh.authorizedKeys.keys` in `machine.nix`
- **Security group**: Ensure port 22 is open from your IP in `infrastructure.nix`
- **Wrong IP**: If `sshProxyJump` or `usePrivateIp` is set, inframan uses `private_ip`
- **Bastion unreachable**: Test `ssh admin@bastion-host` first

### 5. For Colmena failures

```bash
# Inspect the generated hive
cat .inframan/<project>/colmena/hive.nix

# Validate it
cd .inframan/<project>/colmena
colmena eval -f hive.nix -E '{ nodes, ... }: nodes'
```

Common Colmena issues:
- **Build failure on target**: Set `buildOnTarget = false` in `mkRunner` to build locally
- **Disk space**: Target may not have enough disk for compilation
- **Architecture mismatch**: Check `targetSystem` matches the target machine
- **Nix expression errors**: Run `colmena eval` to see Nix evaluation errors

### 6. For Terraform failures

```bash
cd .inframan/<project>/terraform
terraform plan  # See what would change
terraform show   # See current state
```

Common Terraform issues:
- **Missing credentials**: Check `terraform.tfvars` or `TF_VAR_*` env vars
- **Backend errors**: Check S3 bucket exists and is accessible
- **Resource conflicts**: Resource may already exist (use `inframan import`)

## Quick Reference

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| "INFRA_CONFIG_JSON not set" | Running binary directly | Use `nix run .#<project>` |
| "No IP found" | Infra not provisioned | Run `inframan infra` first |
| SSH timeout | Security group / wrong IP | Check SG rules and IP mode |
| Colmena build failure | Disk space / arch mismatch | Set `buildOnTarget = false` or check `targetSystem` |
| "project does not exist" | Typo or not initialized | Check `ls .inframan/` |
