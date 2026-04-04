---
name: add-config-option
description: How to add a new configuration option (environment variable + mkRunner parameter)
trigger: When the user asks to add a new config option, environment variable, or mkRunner parameter
---

# Adding a New Configuration Option

Configuration in inframan flows: Nix (`lib.mkRunner` parameter) → environment variable → Go getter function.

## Steps

### 1. Add getter in `internal/orchestrator/config.go`

**For optional strings:**

```go
// GetMyOption returns MY_OPTION from environment, or empty string if not set
func GetMyOption() string {
    return os.Getenv("MY_OPTION")
}
```

**For strings with a default:**

```go
func GetMyOption() string {
    v := os.Getenv("MY_OPTION")
    if v == "" {
        return "default-value"
    }
    return v
}
```

**For booleans (default true):**

```go
func GetMyFlag() bool {
    v := os.Getenv("MY_FLAG")
    if v == "false" || v == "0" {
        return false
    }
    return true
}
```

**For booleans (default false):**

```go
func GetMyFlag() bool {
    v := os.Getenv("MY_FLAG")
    return v == "true" || v == "1"
}
```

### 2. Add parameter to `lib.mkRunner` in `flake.nix`

Add the parameter with a default in the function signature:

```nix
lib.mkRunner = { system, ..., myOption ? null }:
```

Add the conditional export in the `let` block:

```nix
myOptionExport = if myOption != null
  then ''export MY_OPTION="${myOption}"''
  else "";
```

For booleans:

```nix
lib.mkRunner = { system, ..., myFlag ? true }:
# ...
myFlagExport = if myFlag
  then ''export MY_FLAG="true"''
  else ''export MY_FLAG="false"'';
```

Add the export variable to the `writeShellApplication` text block:

```nix
text = ''
  # ... existing exports
  ${myOptionExport}
  exec ${inframanBin}/bin/inframan "$@"
'';
```

### 3. Use the getter

In the relevant command or orchestrator function:

```go
myOption := orchestrator.GetMyOption()
if myOption != "" {
    // use it
}
```

### 4. Document

1. Add a comment block above the `lib.mkRunner` function in `flake.nix` describing the parameter
2. Add to the environment variables list in `internal/cli/root.go` Long description
3. Add to the README.md environment variables table
4. Add to `docs/configuration.md`

## Reference

- **Optional string pattern**: `GetSSHKeyPath()` in `config.go`
- **Boolean default-true pattern**: `GetBuildOnTarget()` in `config.go`
- **Boolean default-false pattern**: `GetUsePrivateIP()` in `config.go`
- **String with default pattern**: `GetProjectName()`, `GetTargetSystem()` in `config.go`
- **Nix export pattern**: see `sshKeyExport`, `buildOnTargetExport` in `flake.nix`
