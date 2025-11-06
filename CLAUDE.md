# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LazyLinux is a macOS-based Linux VM management tool that provides easy creation and management of QEMU-based virtual machines. The project consists of two main components:

1. **Core CLI Tool** (`./vm`): A bash script that handles VM lifecycle, image management, and QEMU orchestration
2. **VS Code Extension** (`vscode-vm-manager/`): TypeScript-based extension providing GUI management of VMs

## Architecture

### Core CLI (`./vm`)
- **Language**: Bash script
- **Main executable**: `./vm` (single file containing all functionality)
- **Data directory**: `~/.vm/` (images/, vms/ subdirectories)
- **VM formats**: QEMU qcow2 disk images with cloud-init configuration
- **Supported architectures**: ARM64 (native) and AMD64/x86_64 (emulated)
- **Network modes**: Bridge networking and port forwarding
- **Images supported**: Debian 12 (Bookworm) and Debian 13 (Trixie)

### VS Code Extension
- **Language**: TypeScript
- **Structure**: Standard VS Code extension with tree providers for VMs
- **Location**: `vscode-vm-manager/` directory
- **Build**: `npm run compile` compiles TypeScript to JavaScript in `out/`
- **Package**: Uses `vsce package` to create `.vsix` file
- **Communication**: Executes CLI commands via child processes

## Common Development Commands

### VS Code Extension Development
```bash
cd vscode-vm-manager/
npm install                    # Install dependencies
npm run compile               # Build TypeScript to JavaScript
npm run watch                 # Watch mode for development
vsce package                  # Create installable .vsix package
```

### CLI Development
The main CLI is a single bash script (`./vm`) - no build process required. Test by running directly:
```bash
./vm create --name test    # Create test VM
./vm list                  # List VMs
./vm delete test --force   # Clean up
```

### Testing
Run the comprehensive test suite:
```bash
./test/test_linux.sh         # Run all tests
```

## Key Implementation Details

### VM Management Flow
1. **Image Download**: Cloud images downloaded to `~/.vm/images/` (qcow2 format)
2. **VM Creation**: Creates VM directory in `~/.vm/vms/{name}/` with:
   - `disk.qcow2`: VM disk (copy-on-write from base image)
   - `seed.iso`: Cloud-init configuration for SSH setup
   - `info.json`: VM metadata (ports, credentials, etc.)
   - `console.log`: QEMU console output
   - `qemu.pid`: Process ID when running

### Network Configuration
- **Default mode**: Bridge networking (requires bridge networking capability)
- **Fallback mode**: Port forwarding (`--net-type portfwd`)
- **SSH access**: Bridge mode uses host network DHCP, port forwarding uses localhost ports

### Architecture Support
- **ARM64**: Native execution on Apple Silicon (uses qemu-system-aarch64)
- **AMD64**: Emulated execution on Apple Silicon (uses qemu-system-x86_64)
- **Architecture detection**: Automatic based on `uname -m` or explicit via --arch flag

### VS Code Integration
- **Tree Provider**: `vmTreeProvider.ts` manages VM UI tree
- **Script Communication**: `vmScriptUtils.ts` handles CLI command execution
- **Configuration**: Extension settings stored in VS Code workspace/user settings
- **Commands**: Create, start, stop, delete VMs, SSH connection, IP retrieval

## Testing and Quality Assurance

### Manual Testing Commands
```bash
# Test basic VM lifecycle
./vm create --name test-basic
./vm start test-basic
./vm list  # Should show RUNNING
./vm stop test-basic
./vm delete test-basic --force

# Test architecture support
./vm create --name test-amd64 --arch amd64
./vm list  # Verify architecture in output

# Test VS Code extension
cd vscode-vm-manager && npm run compile
vsce package
code --install-extension vm-manager-0.1.0.vsix
```

### Automated Testing
```bash
# Run full test suite
./test/test_linux.sh

# Test suite covers:
# - Command validation and error handling
# - VM lifecycle operations
# - Architecture validation
# - Network configuration
# - Cleanup and purge operations
```

## Development Workflow

1. **Core CLI changes**: Edit `./vm` directly, test with VM lifecycle commands
2. **VS Code extension**:
   - Modify TypeScript files in `vscode-vm-manager/src/`
   - Run `npm run compile`
   - Reload VS Code window to test changes
3. **Testing changes**: Run `./test/test_linux.sh` to validate functionality

## Important File Locations

- **CLI script**: `./vm` (main executable)
- **VS Code extension entry**: `vscode-vm-manager/src/extension.ts`
- **VM data directory**: `~/.vm/` (created by CLI)
- **Extension configuration**: `vscode-vm-manager/package.json` (VS Code integration)
- **Test suite**: `test/test_linux.sh`
- **Tree provider**: `vscode-vm-manager/src/vmTreeProvider.ts`
- **CLI utils**: `vscode-vm-manager/src/vmScriptUtils.ts`

## Troubleshooting Common Issues

### VS Code Extension Not Working
1. Check if `vm` CLI is in PATH or configure `vmManager.scriptPath`
2. Verify extension compiled: `npm run compile` in `vscode-vm-manager/`
3. Check VS Code developer console for errors

### VM Creation Fails
1. Verify QEMU installed: `qemu-system-aarch64 --version`
2. Check disk space in `~/.vm/`
3. Review console logs in `~/.vm/vms/{name}/console.log`
4. Ensure Homebrew is installed (required for automatic QEMU installation)

### Network Issues
1. Try port forwarding mode: `./vm create --name test --net-type portfwd`
2. Check if bridge networking is supported on your system
3. Verify no port conflicts: `lsof -i :PORT_NUMBER`