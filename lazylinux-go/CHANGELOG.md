# Changelog

All notable changes to the LazyLinux VM Manager (Go edition) will be documented in this file.

## [Unreleased]

### Added
- ✅ **cleanup** command - Find and kill orphaned QEMU processes
- ✅ **purge** command - Delete all VMs with confirmation
- ✅ **Multiple VM creation** - Create multiple VMs at once with comma-separated names
  - Example: `vm create --name vm1,vm2,vm3 --user admin --pass test`
  - Auto-generates unique IPs and ports for each VM
  - Shows summary of all created VMs

### Features

#### Cleanup Command
```bash
# Find and kill orphaned QEMU processes
vm cleanup
```

Automatically detects:
- QEMU processes for VMs that no longer exist
- Process ownership (user vs root)
- Prompts for confirmation before killing
- Handles sudo automatically for root-owned processes

#### Purge Command
```bash
# Delete all VMs (with confirmation)
vm purge

# Force delete without confirmation
vm purge --force
```

Features:
- Lists all VMs before deletion
- Shows which VMs are running
- Stops running VMs automatically
- Provides deletion summary

#### Multiple VM Creation
```bash
# Create 3 VMs with same configuration
vm create --name web1,web2,web3 --memory 8G --cpus 4

# Create with shared credentials
vm create --name db1,db2 --user admin --pass mypass

# Mix with other flags
vm create --name vm1,vm2,vm3 --net-type portfwd
```

Features:
- Auto-generates unique static IPs (bridge mode)
- Auto-assigns unique SSH ports (portfwd mode)
- Shares username/password across all VMs
- Progress indicator for each VM
- Summary table with all connection details
- Continues on failure (doesn't stop if one VM fails)

## [0.1.0] - Initial Release

### Core Features
- VM creation with QEMU
- Bridge and port forwarding networking
- Cloud-init configuration
- VM lifecycle management (create, start, stop, delete)
- SSH access
- IP detection
- Diagnostics
- Architecture support (ARM64, AMD64)

### Commands
- `vm create` - Create VMs
- `vm list` - List VMs
- `vm start` - Start VM
- `vm stop` - Stop VM
- `vm delete` - Delete VM
- `vm ssh` - SSH to VM
- `vm ip` - Get VM IP
- `vm setup-bridge` - Setup bridge networking
- `vm diagnose` - Run diagnostics

### Documentation
- Complete README
- Migration guide from bash version
- Quick start guide
- Configuration examples
