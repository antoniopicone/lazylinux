# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LazyLinux is a macOS-based Linux VM management tool that provides easy creation and management of QEMU-based virtual machines. The project consists of three main components:

1. **Core CLI Tool** (`./linux`): A bash script that handles VM lifecycle, image management, and QEMU orchestration
2. **VS Code Extension** (`vscode-extension/`): TypeScript-based extension providing GUI management of VMs
3. **VMNet Bridge Helper** (`VMNetBridge/`): Swift-based privileged helper for bridged networking

## Architecture

### Core CLI (`./linux`)
- **Language**: Bash script
- **Main executable**: `./linux` (single file containing all functionality)
- **Data directory**: `~/.linux/` (images/, vms/ subdirectories)
- **VM formats**: QEMU qcow2 disk images with cloud-init configuration
- **Supported architectures**: ARM64 (native) and AMD64/x86_64 (emulated)
- **Network modes**: Bridge networking (via VMNetBridge) and port forwarding

### VS Code Extension
- **Language**: TypeScript
- **Structure**: Standard VS Code extension with tree providers for VMs and images
- **Build**: `npm run compile` compiles TypeScript to JavaScript in `out/`
- **Package**: Uses `vsce package` to create `.vsix` file
- **Communication**: Executes CLI commands via child processes

### VMNet Bridge Helper
- **Language**: Swift
- **Build system**: Swift Package Manager
- **Target**: Privileged launchd daemon for bridged networking
- **Installation**: Requires root privileges, installs to `/usr/local/bin/` and `/Library/LaunchDaemons/`

## Common Development Commands

### VS Code Extension Development
```bash
cd vscode-extension/
npm install                    # Install dependencies
npm run compile               # Build TypeScript to JavaScript
npm run watch                 # Watch mode for development
vsce package                  # Create installable .vsix package
```

### VMNet Bridge Development
```bash
cd VMNetBridge/
swift build -c release       # Build release binary
swift run vmnet-helper        # Run in development mode
./install.sh                 # Install as system service (requires sudo)
./uninstall.sh               # Remove system service
```

### CLI Development
The main CLI is a single bash script (`./linux`) - no build process required. Test by running directly:
```bash
./linux vm create --name test    # Create test VM
./linux vm list                  # List VMs
./linux vm delete test --force   # Clean up
```

## Key Implementation Details

### VM Management Flow
1. **Image Download**: Cloud images downloaded to `~/.linux/images/` (qcow2 format)
2. **VM Creation**: Creates VM directory in `~/.linux/vms/{name}/` with:
   - `disk.qcow2`: VM disk (copy-on-write from base image)
   - `seed.iso`: Cloud-init configuration for SSH setup
   - `info.json`: VM metadata (ports, credentials, etc.)
   - `console.log`: QEMU console output
   - `qemu.pid`: Process ID when running

### Network Configuration
- **Default mode**: VMNet shared bridge networking (no sudo required after helper installation)
- **Fallback mode**: Port forwarding (--net-type portfwd)
- **Bridge mode**: Uses macOS VMNet framework for direct host network access
- **SSH access**: Bridge mode uses host network DHCP, port forwarding uses localhost ports

### Architecture Support
- **ARM64**: Native execution on Apple Silicon (uses qemu-system-aarch64)
- **AMD64**: Emulated execution on Apple Silicon (uses qemu-system-x86_64)
- **Architecture detection**: Automatic based on `uname -m` or explicit via --arch flag

### VS Code Integration
- **Tree Providers**: `vmTreeProvider.ts` and `imageTreeProvider.ts` manage UI trees
- **Script Communication**: `linuxScriptUtils.ts` handles CLI command execution
- **Configuration**: Extension settings stored in VS Code workspace/user settings
- **Auto-refresh**: Configurable polling of VM status

## Testing and Quality Assurance

### Manual Testing Commands
```bash
# Test basic VM lifecycle
./linux vm create --name test-basic
./linux vm start test-basic
./linux vm list  # Should show RUNNING
./linux vm stop test-basic
./linux vm delete test-basic --force

# Test architecture support  
./linux vm create --name test-amd64 --arch amd64
./linux vm list  # Verify architecture in output

# Test image management
./linux image list
./linux image pull debian12-arm64
./linux image delete debian12-arm64 --force

# Test VS Code extension
cd vscode-extension && npm run compile
code --install-extension linux-vm-manager-0.1.0.vsix
```

### VMNet Bridge Testing
```bash
# Test service installation
cd VMNetBridge && sudo ./install.sh
sudo launchctl list | grep vmnet-helper

# Verify helper is running
ls -la /tmp/vmnet-bridge-ready  # Should exist

# Test bridge mode VM
./linux vm create --name bridge-test --net-type bridge
./linux vm list  # Should show bridge-test with host-network-ip
```

## Development Workflow

1. **Core CLI changes**: Edit `./linux` directly, test with VM lifecycle commands
2. **VS Code extension**: 
   - Modify TypeScript files in `vscode-extension/src/`
   - Run `npm run compile`
   - Reload VS Code window to test changes
3. **VMNet Bridge**: 
   - Edit Swift files in `VMNetBridge/Sources/`
   - Build and test with `swift run vmnet-helper`
   - Install system service for full integration testing

## Important File Locations

- **CLI script**: `./linux` (main executable)
- **VS Code extension entry**: `vscode-extension/src/extension.ts`
- **VM data directory**: `~/.linux/` (created by CLI)
- **VMNet helper**: `VMNetBridge/Sources/VMNetHelper/main.swift`
- **Extension configuration**: `vscode-extension/package.json` (VS Code integration)
- **Swift package**: `VMNetBridge/Package.swift`

## Troubleshooting Common Issues

### VS Code Extension Not Working
1. Check if `linux` CLI is in PATH or configure `linuxVmManager.scriptPath`
2. Verify extension compiled: `npm run compile` in `vscode-extension/`
3. Check VS Code developer console for errors

### VM Creation Fails
1. Verify QEMU installed: `qemu-system-aarch64 --version`
2. Check disk space in `~/.linux/`
3. Review console logs in `~/.linux/vms/{name}/console.log`

### Bridge Networking Issues
1. Verify VMNet helper service: `sudo launchctl list | grep vmnet-helper`
2. Check helper status marker: `ls -la /tmp/vmnet-bridge-ready`
3. Verify QEMU VMNet support: `qemu-system-aarch64 -netdev help | grep vmnet`
4. Check helper logs: `tail -f /var/log/vmnet-helper.log`