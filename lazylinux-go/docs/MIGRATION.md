# Migration Guide: Bash to Go

This guide helps you migrate from the bash version of LazyLinux VM Manager to the Go version.

## Compatibility

The Go version is **fully compatible** with VMs created by the bash version. Both versions:

- Use the same directory structure (`~/.vm/`)
- Use the same `info.json` format
- Use the same QEMU configurations
- Can manage the same VMs

You can use both versions side-by-side without issues.

## Command Comparison

| Bash Version | Go Version | Notes |
|--------------|------------|-------|
| `./vm create` | `vm create` | Same options available |
| `./vm list` | `vm list` | Same output format |
| `./vm start <name>` | `vm start <name>` | Identical |
| `./vm stop <name>` | `vm stop <name>` | Identical |
| `./vm delete <name>` | `vm delete <name>` | Use `--force` flag |
| `./vm ssh <name>` | `vm ssh <name>` | Identical |
| `./vm ip <name>` | `vm ip <name>` | Identical |
| `./vm setup-bridge` | `vm setup-bridge` | Identical |
| `./vm diagnose` | `vm diagnose` | Improved output |
| `./vm install` | `make install` | Installation method |
| `./vm cleanup` | `vm cleanup` | Identical |
| `./vm purge` | `vm purge` | Use `--force` flag |

## Installation

### Uninstall Bash Version (Optional)

```bash
# If you installed the bash version system-wide
sudo rm /usr/local/bin/vm

# The bash script itself can remain for reference
```

### Install Go Version

```bash
cd lazylinux-go
make build
make install
```

Or manually:

```bash
go build -o vm ./cmd/vm
sudo mv vm /usr/local/bin/
```

## Configuration

The Go version introduces optional configuration via `~/.vm/config.yaml`:

```yaml
work_root: ~/.vm
images_dir: ~/.vm/images
vms_dir: ~/.vm/vms
default_username: user01
default_memory: 4G
default_cpus: 2
default_disk_size: 10G
default_image: debian13
```

This file is **optional**. If it doesn't exist, sensible defaults are used.

## Feature Parity

### Implemented Features

✅ VM creation with all options
✅ VM lifecycle (start, stop, delete)
✅ Network modes (bridge, port forwarding)
✅ SSH access
✅ IP detection
✅ Bridge networking setup
✅ Diagnostics
✅ Auto-generated VM names and passwords
✅ Static IP assignment
✅ Multiple architectures (ARM64, AMD64)

### Recently Added Features

✅ Cleanup orphaned QEMU processes
✅ Purge all VMs
✅ Multiple VM creation (comma-separated names)

### Not Yet Implemented

⏳ Show console mode
⏳ Install VS Code extension

These features will be added in future versions.

## Differences

### Improved Error Handling

The Go version provides more detailed error messages:

```bash
# Bash version
ERROR: VM 'myvm' not found

# Go version
Error: VM not found: failed to read info file: open ~/.vm/vms/myvm/info.json: no such file or directory
```

### Better Performance

- Faster startup time (compiled binary vs interpreted script)
- Concurrent operations where possible
- Efficient file I/O

### Enhanced Diagnostics

The `diagnose` command provides more structured output:

```bash
vm diagnose
```

Output includes:
- QEMU installation status
- Bridge networking status
- Running VMs count
- System dependencies

### Structured Logging

More consistent output formatting:

```bash
# Creating a VM
Downloading base image debian13 for arm64...
Creating VM disk...
Generating cloud-init configuration...
Starting VM...

✓ VM 'myvm' created successfully!

Credentials: user01 / abc123def456
SSH: ssh user01@192.168.105.150
```

## Testing the Migration

After installing the Go version:

1. **List existing VMs:**
   ```bash
   vm list
   ```

2. **Start an existing VM:**
   ```bash
   vm start <existing-vm-name>
   ```

3. **SSH into existing VM:**
   ```bash
   vm ssh <existing-vm-name>
   ```

4. **Create a new VM:**
   ```bash
   vm create --name test-go
   ```

5. **Compare with bash version:**
   ```bash
   ./vm list  # Should show same VMs
   ```

## Troubleshooting

### "Command not found: vm"

Make sure `/usr/local/bin` is in your PATH:

```bash
echo $PATH | grep "/usr/local/bin"
```

If not, add to your `~/.zshrc` or `~/.bashrc`:

```bash
export PATH="/usr/local/bin:$PATH"
```

### "Permission denied"

The binary needs execute permissions:

```bash
chmod +x /usr/local/bin/vm
```

### "Cannot find QEMU"

Install QEMU:

```bash
brew install qemu
```

Or let the tool install it automatically when you run a command.

### VMs created with bash version don't work

This shouldn't happen. If it does:

1. Check VM info file:
   ```bash
   cat ~/.vm/vms/<vm-name>/info.json
   ```

2. Verify QEMU process:
   ```bash
   pgrep -f "qemu.*-name <vm-name>"
   ```

3. Check console log:
   ```bash
   tail ~/.vm/vms/<vm-name>/console.log
   ```

## Rollback

If you need to go back to the bash version:

```bash
# Remove Go binary
sudo rm /usr/local/bin/vm

# Reinstall bash version
cd /path/to/lazylinux
./vm install
```

Your VMs remain untouched and will work with either version.

## Getting Help

- **View all commands:** `vm --help`
- **Command-specific help:** `vm <command> --help`
- **Diagnostics:** `vm diagnose`
- **GitHub Issues:** Report bugs or request features

## Future Improvements

The Go version is actively developed. Planned improvements:

- [ ] Web UI for VM management
- [ ] VM templates
- [ ] Snapshot support
- [ ] VM cloning
- [ ] Resource monitoring
- [ ] Batch operations
- [ ] Configuration profiles
- [ ] Export/import VMs

## Feedback

We'd love to hear your experience migrating to the Go version:

- What works well?
- What's missing?
- What could be improved?

Please open a GitHub issue or discussion!
