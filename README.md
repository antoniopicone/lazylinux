# Linux VM CLI

A simple command-line tool for creating and managing ARM64 Linux virtual machines on macOS using QEMU.

## Features

- **Easy VM Management**: Create, start, stop, and delete VMs with simple commands
- **Image Management**: Download and manage cloud images (Debian 12, for now)
- **Automatic Setup**: Cloud-init configuration with SSH access
- **Port Management**: Automatic SSH port assignment with conflict resolution
- **Persistent Storage**: VM state preserved across restarts
- **macOS Optimized**: Uses HVF acceleration for better performance

## Installation

### Prerequisites

**Homebrew is required** - The script will automatically install QEMU if not available, but needs Homebrew to do so.

Install Homebrew if you don't have it:
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

### Install the Linux CLI

```bash
# Download, make executable, and install
curl -o linux https://raw.githubusercontent.com/antoniopicone/lazylinux/refs/heads/main/linux
chmod +x linux
sudo mv linux /usr/local/bin/
```

Verify installation:
```bash
linux --help
```

## Quick Start

### 1. Create your first VM

```bash
# Create a VM with default settings (Debian 12, 2GB RAM, 2 CPUs)
linux vm create --name myvm

# Create a VM with custom specifications
linux vm create --name myvm --memory 4G --cpus 4 --disk 20G
```

### 2. Manage VMs

```bash
# List all VMs
linux vm list

# Start a VM
linux vm start myvm

# Stop a VM
linux vm stop myvm

# Delete a VM
linux vm delete myvm
```

### 3. Manage Images

```bash
# List available images
linux image list

# Delete a downloaded image
linux image delete debian12
```

## Usage

### VM Commands

```bash
# Create and start a new VM
linux vm create --name NAME [options]

# List all VMs and their status  
linux vm list

# Start an existing VM
linux vm start VM_NAME

# Stop a running VM
linux vm stop VM_NAME

# Delete a VM and its files
linux vm delete VM_NAME
```

### VM Creation Options

| Option | Description | Default |
|--------|-------------|---------|
| `--name NAME` | VM name (required) | - |
| `--image IMAGE` | Base image | debian12 |
| `--user USERNAME` | Username | user01 |
| `--pass PASSWORD` | Password | auto-generated |
| `--memory SIZE` | Memory size | 2G |
| `--cpus COUNT` | CPU count | 2 |
| `--disk SIZE` | Disk size | 10G |
| `--ssh-port PORT` | SSH port | auto-assign |
| `--show-console` | Show console output | false |

### Image Commands

```bash
# List available and downloaded images
linux image list

# Download an image
linux image pull IMAGE_NAME

# Delete a downloaded image  
linux image delete IMAGE_NAME
```

### Available Images

| Image | Description |
|-------|-------------|
| `debian12` | Debian 12 (Bookworm) ARM64 |

## Examples

### Basic VM Creation

```bash
# Create a simple VM (uses Debian 12 by default)
linux vm create --name dev

# Create a VM with custom credentials
linux vm create --name secure --user admin --pass mypassword
```

### Working with VMs

```bash
# Check VM status
linux vm list

# Start with console access (useful for debugging)
linux vm start myvm --show-console

# Connect via SSH (get details from 'linux vm list')
ssh user01@127.0.0.1 -p 2222
```

### Image Management

```bash
# Check downloaded images
linux image list

# Clean up unused images
linux image delete debian12
```

## Directory Structure

The CLI stores all data in `~/.linux_vm_cli/`:

```
~/.linux_vm_cli/
├── images/          # Downloaded base images
│   ├── debian12.qcow2
│   ├── ubuntu22.qcow2
│   └── ubuntu24.qcow2
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
   linux vm list  # Shows assigned ports
   lsof -i :2222  # Check specific port
   ```

2. View console logs:
   ```bash
   cat ~/.linux_vm_cli/vms/myvm/console.log
   ```

3. Start with console output for debugging:
   ```bash
   linux vm start myvm --show-console
   ```

### SSH Connection Issues

1. Ensure the VM is fully booted:
   ```bash
   linux vm list  # Check status is RUNNING
   ```

2. Test SSH connectivity:
   ```bash
   nc -z 127.0.0.1 2222  # Replace 2222 with your VM's port
   ```

3. Check VM console for cloud-init completion:
   ```bash
   grep "CLOUD-INIT-READY" ~/.linux_vm_cli/vms/myvm/console.log
   ```

## Requirements

- macOS (Intel or Apple Silicon)
- Homebrew (for automatic QEMU installation)
- Internet connection (for downloading images)

## License

MIT License - feel free to modify and distribute.

## Contributing

Issues and pull requests welcome! Please ensure any changes maintain compatibility with both Intel and Apple Silicon Macs.