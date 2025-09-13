# VM Manager - VS Code Extension

A VS Code extension for managing Linux VMs created with the `vm` script from LazyLinux. This extension provides a visual interface to view, create, start, stop, and manage virtual machines directly from VS Code.

**Publisher**: Antonio Picone
**Repository**: [antoniopicone/lazylinux](https://github.com/antoniopicone/lazylinux)

## About LazyLinux

This extension is the VS Code frontend for the **LazyLinux** project - a macOS-based Linux VM management tool that provides easy creation and management of QEMU-based virtual machines. The `vm` script is the core CLI tool that handles VM lifecycle, image management, and QEMU orchestration.

## Features

- See all your VMs in the Explorer panel with real-time status and VM details showing hostname, username, and password (hidden)
- Create VMs with a guided setup**: Name → Image → Architecture → Username → Password
- Multiple architecture supported: arm64 (native) and amd64 (emulated) with performance hints
- SSH Integration with direct terminal connection with parsed credentials


## Installation

1. Make sure you have the LazyLinux `vm` script installed:
   ```bash
   git clone https://github.com/antoniopicone/lazylinux.git
   cd lazylinux
   ./vm install  # Optional: Install to system PATH
   ```

2. Install the VS Code extension from the VSIX file:
   ```bash
   code --install-extension vm-manager-0.1.0.vsix
   ```

3. Open VS Code and look for the "Virtual Machines" section in the Explorer panel

## Configuration

Configure the path to your `vm` script in VS Code settings:

```json
{
    "vmManager.scriptPath": "/usr/local/bin/vm"
}
```

**Default paths tried**:
- `/usr/local/bin/vm` (if installed to system)
- `./vm` (in current workspace)

## Usage

### 🚀 Creating VMs
1. Click the "+" icon in the VM Manager header
2. **Step 1/5**: Enter VM name (or leave empty for auto-generated)
3. **Step 2/5**: Select base image (debian13 recommended)
4. **Step 3/5**: Choose architecture (arm64 for Apple Silicon)
5. **Step 4/5**: Set username (defaults to "user01")
6. **Step 5/5**: Set password (leave blank for auto-generated)


## Requirements

- **VS Code**: 1.74.0 or higher
- **LazyLinux vm script**: Available from [antoniopicone/lazylinux](https://github.com/antoniopicone/lazylinux)

## Architecture Support

- **ARM64**: Native execution on Apple Silicon (recommended)
- **AMD64/x86_64**: Emulated execution on Apple Silicon

## Network Modes

The extension supports LazyLinux network configurations:
- **Bridge Mode**: Direct host network access (default)
- **Port Forwarding**: Localhost port mapping

## Development

### Building from Source
```bash
cd vscode-vm-manager
npm install
npm run compile
vsce package
```

### Development Mode
```bash
npm run watch  # Auto-compilation during development
```

### Testing
```bash
# Install locally for testing
code --install-extension vm-manager-0.1.0.vsix
```

## Extension Structure

- **`src/extension.ts`**: Main extension entry point and command registration
- **`src/vmTreeProvider.ts`**: Tree view provider with credential parsing
- **`src/vmScriptUtils.ts`**: LazyLinux vm script integration and VM operations
- **`package.json`**: Extension manifest with commands and configuration

## Supported VM Images

The extension supports all LazyLinux-compatible images:
- **debian13** (default)
- **debian12**

## Contributing

This extension is part of the LazyLinux project. Contributions welcome!

**Repository**: [https://github.com/antoniopicone/lazylinux](https://github.com/antoniopicone/lazylinux)
**Issues**: Report bugs and feature requests in the main LazyLinux repository
**License**: Check the LazyLinux repository for licensing information

## Author

**Antonio Picone**
- GitHub: [@antoniopicone](https://github.com/antoniopicone)
- Project: [LazyLinux](https://github.com/antoniopicone/lazylinux)

---

*This VS Code extension provides an user-friendly interface for LazyLinux VM management system.*