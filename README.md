# VM Manager

A simple command-line tool for creating and managing Linux virtual machines on macOS using QEMU.

## Features

- **Easy VM Management**: Create, start, stop, and delete VMs with simple commands
- **Image Management**: Download and manage cloud images (Debian 13 and 12, for now)
- **Virtualization & Emulation**: Support both ARM and AMD architectures
- **Automatic Setup**: Cloud-init configuration with SSH access
- **Port Management**: Automatic SSH port assignment with conflict resolution
- **Persistent Storage**: VM state preserved across restarts
- **macOS Optimized**: Uses HVF acceleration for better performance
- **VS Code Integration**: Manage VMs directly from Visual Studio Code with the LazyLinux extension

## Installation

### Prerequisites

**Homebrew is required** - The script will automatically install QEMU if not available, but needs Homebrew to do so.

Install Homebrew if you don't have it:
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

### Install the VM CLI

```bash
# Download, make executable, and install
curl -o vm https://raw.githubusercontent.com/antoniopicone/lazylinux/refs/heads/main/vm
chmod +x vm

# Install to system PATH (optional)
./vm install
```

Verify installation:
```bash
vm help
```

## Quick Start

### 1. Create your first VM

```bash
# Create a VM with auto-generated name (Debian 13, 2GB RAM, 2 CPUs)
vm create

# Create a VM with custom name and specifications
vm create --name myvm --memory 4G --cpus 4 --disk 20G

# Create an x86_64 (amd64) VM (emulated on Apple Silicon)
vm create --name myvm-amd64 --arch amd64
```

### 2. Manage VMs

```bash
# List all VMs
vm list

# Start a VM
vm start myvm

# Stop a VM
vm stop myvm

# Delete a VM
vm delete myvm
```

### 3. System operations

```bash
# Install vm command to system PATH
vm install

# Remove vm command from system PATH  
vm uninstall
```

### 4. System cleanup

```bash
# Purge all VMs (interactive confirmation)
vm purge

# Purge without confirmation
vm purge --force
```

## Usage

### VM Commands

```bash
# Create and start a new VM
vm create [--name NAME] [options]

# List all VMs and their status  
vm list

# Start an existing VM
vm start VM_NAME

# Stop a running VM
vm stop VM_NAME

# Delete a VM and its files
vm delete VM_NAME
```

### VM Creation Options

| Option | Description | Default |
|--------|-------------|---------|
| `--name NAME` | VM name (optional) | auto-generated |
| `--arch ARCH` | Architecture (`arm64`, `amd64`) | arm64 |
| `--user USERNAME` | Username | user01 |
| `--pass PASSWORD` | Password | auto-generated |
| `--memory SIZE` | Memory size | 2G |
| `--cpus COUNT` | CPU count | 2 |
| `--disk SIZE` | Disk size | 10G |
| `--ssh-port PORT` | SSH port | auto-assign |
| `--net-type TYPE` | Network type (`bridge`, `portfwd`) | bridge |
| `--show-console` | Show console output | false |

### System Commands

```bash
# Install vm command to system PATH
vm install

# Remove vm command from system PATH
vm uninstall

# Get help
vm help
```

### Maintenance Commands

```bash
# Delete ALL VMs (asks for confirmation)
vm purge

# Force purge without confirmation
vm purge --force
```

### Available Images

| Image | Description |
|-------|-------------|
| `debian12` | Debian 12 (Bookworm) |
| `debian13` | Debian 13 (Trixie) |

## Examples

### Basic VM Creation

```bash
# Create a simple VM with auto-generated name (uses Debian 13 by default)
vm create

# Create a VM with specific name
vm create --name dev

# Create a VM with custom credentials
vm create --name secure --user admin --pass mypassword

# Create an amd64 VM instead of the default arm64
vm create --name dev-amd64 --arch amd64
```

### Working with VMs

```bash
# Check VM status
vm list

# Start with console access (useful for debugging)
vm start myvm --show-console

# Connect via SSH (get details from 'vm list')
# For bridge mode VMs:
ssh user01@myvm.local

# For port forwarding VMs:
ssh user01@127.0.0.1 -p 2222
```

### VM Management

```bash
# Connect to a VM directly
vm ssh myvm

# Get VM IP address (for bridge mode)
vm ip myvm
```

### System Maintenance

```bash
# Completely reset the CLI state (remove all VMs)
vm purge --force

# Install/uninstall system command
vm install      # Install to /usr/local/bin
vm uninstall    # Remove from PATH
```

## Directory Structure

The CLI stores all data in `~/.vm/`:

```
~/.vm/
├── images/          # Downloaded base images
│   └── debian13-arm64.qcow2
└── vms/            # VM instances
    └── myvm/
        ├── disk.qcow2      # VM disk
        ├── seed.iso        # Cloud-init configuration
        ├── info.json       # VM metadata
        ├── console.log     # Console output
        └── qemu.pid        # Process ID when running
```

## Troubleshooting

### QEMU Installation Issues

If automatic installation fails:
```bash
# Install QEMU manually
brew install qemu

# Verify installation
qemu-system-aarch64 --version
```

### VM Won't Start

1. Check if another process is using the SSH port:
   ```bash
   vm list        # Shows assigned ports
   lsof -i :2222  # Check specific port
   ```

2. View console logs:
   ```bash
   cat ~/.vm/vms/myvm/console.log
   ```

3. Start with console output for debugging:
   ```bash
   vm start myvm --show-console
   ```

### SSH Connection Issues

1. Ensure the VM is fully booted:
   ```bash
   vm list        # Check status is RUNNING
   ```

2. Test SSH connectivity:
   ```bash
   nc -z 127.0.0.1 2222  # Replace 2222 with your VM's port
   ```

3. Check VM console for cloud-init completion:
   ```bash
   grep "CLOUD-INIT-READY" ~/.vm/vms/myvm/console.log
   ```

## VS Code Extension

A VS Code extension for managing Linux VMs created with the `vm` script from LazyLinux. This extension provides a visual interface to view, create, start, stop, and manage virtual machines directly from VS Code.


### Features

- See all your VMs in the Explorer panel with real-time status and VM details showing hostname, username, and password (hidden)
- Create VMs with a guided setup: Name → Image → Architecture → Username → Password
- Multiple architecture supported: arm64 (native) and amd64 (emulated) with performance hints
- SSH Integration with direct terminal connection with parsed credentials

### Installation

1. **Install the CLI first** (see installation instructions above)

2. **Install VS Code Extension**:
   ```bash
   # Clone or navigate to the vscode-vm-manager directory
   cd vscode-vm-manager

   # Install dependencies and compile
   npm install
   npm run compile

   # Package the extension (optional, for distribution)
   vsce package

   # Install the extension locally
   code --install-extension vm-manager-0.1.0.vsix
   ```

3. **Open VS Code** and look for the "Virtual Machines" section in the Explorer panel

### Configuration

Configure the path to your `vm` script in VS Code settings:

```json
{
    "vmManager.scriptPath": "/usr/local/bin/vm"
}
```

**Default paths tried**:
- `/usr/local/bin/vm` (if installed to system)
- `./vm` (in current workspace)

### Usage

#### 🚀 Creating VMs
1. Click the "+" icon in the VM Manager header
2. **Step 1/5**: Enter VM name (or leave empty for auto-generated)
3. **Step 2/5**: Select base image (debian13 recommended)
4. **Step 3/5**: Choose architecture (arm64 for Apple Silicon)
5. **Step 4/5**: Set username (defaults to "user01")
6. **Step 5/5**: Set password (leave blank for auto-generated)

### Requirements

- **VS Code**: 1.74.0 or higher
- **LazyLinux vm script**: Available from [antoniopicone/lazylinux](https://github.com/antoniopicone/lazylinux)

### Architecture Support

- **ARM64**: Native execution on Apple Silicon (recommended)
- **AMD64/x86_64**: Emulated execution on Apple Silicon

### Network Modes

The extension supports LazyLinux network configurations:
- **Bridge Mode**: Direct host network access (default)
- **Port Forwarding**: Localhost port mapping

### Development

#### Building from Source
```bash
cd vscode-vm-manager
npm install
npm run compile
vsce package
```

#### Development Mode
```bash
npm run watch  # Auto-compilation during development
```

#### Testing
```bash
# Install locally for testing
code --install-extension vm-manager-0.1.0.vsix
```

### Extension Structure

- **`src/extension.ts`**: Main extension entry point and command registration
- **`src/vmTreeProvider.ts`**: Tree view provider with credential parsing
- **`src/vmScriptUtils.ts`**: LazyLinux vm script integration and VM operations
- **`package.json`**: Extension manifest with commands and configuration

### Supported VM Images

The extension supports all LazyLinux-compatible images:
- **debian13** (default)
- **debian12**

## Requirements

- macOS (Intel or Apple Silicon)
- Homebrew (for automatic QEMU installation)
- Internet connection (for downloading images)
- Visual Studio Code (optional, for the extension)

## License

MIT License - feel free to modify and distribute.

## Contributing

Issues and pull requests welcome! Please ensure any changes maintain compatibility with both Intel and Apple Silicon Macs.