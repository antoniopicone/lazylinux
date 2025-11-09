# LazyLinux VM Manager (Go Edition)

A fast, modern, and easy-to-use virtual machine manager for macOS, built with Go.

## Features

- ✨ **Simple CLI**: Intuitive command-line interface powered by Cobra
- 🚀 **Fast**: Written in Go for optimal performance
- 🔧 **Easy Setup**: Automatic QEMU installation via Homebrew
- 🌐 **Flexible Networking**: Bridge or port forwarding modes
- 💾 **Cloud Images**: Uses official Debian cloud images
- 🔐 **Secure**: Auto-generated passwords, cloud-init configuration
- 📦 **Modular**: Clean architecture with separation of concerns

## Architecture

```
lazylinux-go/
├── cmd/vm/                 # CLI application
│   ├── commands/           # Cobra commands
│   └── main.go
├── internal/              # Private application code
│   ├── cloudinit/         # Cloud-init generation
│   ├── config/            # Configuration management
│   ├── network/           # Network (bridge) management
│   ├── qemu/              # QEMU operations
│   └── vm/                # VM lifecycle management
├── pkg/                   # Public libraries
│   ├── models/            # Domain models
│   └── utils/             # Utilities
└── templates/             # External templates
    ├── cloud-init-bridge.yaml
    └── cloud-init-portfwd.yaml
```

## Installation

### Prerequisites

- macOS (Apple Silicon or Intel)
- [Homebrew](https://brew.sh/)
- Go 1.21+ (for building from source)

### Building from Source

```bash
cd lazylinux-go
go mod download
go build -o vm ./cmd/vm
sudo mv vm /usr/local/bin/
```

### Setup Bridge Networking (Optional)

For VMs to get their own IP addresses on your local network:

```bash
vm setup-bridge
```

This will install and configure `socket_vmnet` for bridge networking.

## Quick Start

```bash
# Create a VM with auto-generated name and configuration
vm create

# Create a VM with specific settings
vm create --name myvm --memory 8G --cpus 4

# Create a VM with port forwarding (no bridge setup needed)
vm create --name dev --net-type portfwd

# List all VMs
vm list

# SSH into a VM
vm ssh myvm

# Stop a VM
vm stop myvm

# Delete a VM
vm delete myvm

# Get diagnostics
vm diagnose
```

## Commands

### `vm cleanup`

Find and kill orphaned QEMU processes.

```bash
vm cleanup
```

This command:
- Scans for QEMU processes
- Identifies VMs that no longer exist
- Shows process ownership (user vs root)
- Asks for confirmation before killing
- Handles sudo automatically for root processes

### `vm purge`

Delete all VMs and their files.

```bash
# With confirmation
vm purge

# Force delete without confirmation
vm purge --force
```

**WARNING**: This deletes ALL VMs. Use with caution!

### `vm create`

Create a new virtual machine.

**Flags:**
- `--name` - VM name (auto-generated if not specified)
- `--arch` - Architecture: `arm64` (default) or `amd64`
- `--user` - Username (default: `user01`)
- `--pass` - Password (auto-generated if not specified)
- `--memory` - Memory size (default: `4G`)
- `--cpus` - CPU count (default: `2`)
- `--disk` - Disk size (default: `10G`)
- `--net-type` - Network type: `bridge` (default) or `portfwd`
- `--ip` - Static IP for bridge mode (auto-generated if not specified)
- `--ssh-port` - SSH port for portfwd mode
- `--debug` - Show debug information

**Examples:**
```bash
# Simple create with defaults
vm create

# Create with specific name and resources
vm create --name webserver --memory 8G --cpus 4 --disk 20G

# Create with static IP
vm create --name db --ip 192.168.105.200

# Create with port forwarding
vm create --name dev --net-type portfwd --ssh-port 2222

# Create multiple VMs at once
vm create --name vm1,vm2,vm3

# Create multiple VMs with shared config
vm create --name web1,web2,web3 --user admin --pass mypass --memory 8G
```

### `vm list`

List all virtual machines and their status.

```bash
vm list
```

Output example:
```
NAME                 STATUS     ARCH     IMAGE      IP ADDRESS                CREDENTIALS
----                 ------     ----     -----      ----------                -----------
myvm                 RUNNING    arm64    debian13   192.168.105.150           user01 / abc123def456
dev                  RUNNING    arm64    debian13   127.0.0.1:2222            user01 / xyz789ghi012
test                 STOPPED    arm64    debian13   -                         -

Total VMs: 3
```

### `vm start <name>`

Start an existing VM.

```bash
vm start myvm
```

### `vm stop <name>`

Stop a running VM.

```bash
vm stop myvm
```

### `vm delete <name>`

Delete a VM and its files.

```bash
# With confirmation
vm delete myvm

# Force delete without confirmation
vm delete myvm --force
```

### `vm ssh <name>`

SSH into a running VM.

```bash
vm ssh myvm
```

### `vm ip <name>`

Get the IP address of a VM.

```bash
vm ip myvm
```

### `vm setup-bridge`

Setup bridge networking with socket_vmnet.

```bash
vm setup-bridge
```

### `vm diagnose`

Run system diagnostics.

```bash
vm diagnose
```

## Configuration

Configuration is stored in `~/.vm/config.yaml`. The default configuration:

```yaml
work_root: ~/.vm
images_dir: ~/.vm/images
vms_dir: ~/.vm/vms
default_username: user01
default_memory: 4G
default_cpus: 2
default_disk_size: 10G
default_image: debian13
brew_prefix: /opt/homebrew

images:
  - name: debian13
    display_name: Debian 13 (Trixie)
    urls:
      arm64: https://cloud.debian.org/images/cloud/trixie/latest/debian-13-genericcloud-arm64.qcow2
      amd64: https://cloud.debian.org/images/cloud/trixie/latest/debian-13-genericcloud-amd64.qcow2
  - name: debian12
    display_name: Debian 12 (Bookworm)
    urls:
      arm64: https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-genericcloud-arm64.qcow2
      amd64: https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-generic-amd64.qcow2
```

## Networking

### Bridge Mode (Default)

VMs get their own IP addresses on your local network using `socket_vmnet`.

**Advantages:**
- VMs accessible from other devices on your network
- No port conflicts
- More realistic network environment

**Requirements:**
- Run `vm setup-bridge` once
- Ethernet connection recommended (WiFi has limitations on macOS)

### Port Forwarding Mode

VMs accessible via localhost ports.

**Advantages:**
- No special setup required
- Works with WiFi
- Isolated from local network

**Usage:**
```bash
vm create --net-type portfwd
ssh user01@127.0.0.1 -p <assigned-port>
```

## Architecture Support

### ARM64 (Apple Silicon)

- Native execution with hardware acceleration (HVF)
- Fast performance
- Default architecture

### AMD64/x86_64

- Emulated on Apple Silicon
- Slower than ARM64
- Use when you need x86_64 compatibility

```bash
vm create --arch amd64
```

## Cloud-Init Templates

Templates are stored in `templates/` directory:

- `cloud-init-bridge.yaml` - Template for bridge networking
- `cloud-init-portfwd.yaml` - Template for port forwarding

These templates use Go's `text/template` syntax and are embedded in the binary at compile time.

## Development

### Project Structure

The project follows Go best practices with a clear separation between:

- **`cmd/`**: Entry points for binaries
- **`internal/`**: Private application code
- **`pkg/`**: Public libraries that could be imported
- **`templates/`**: External configuration templates

### Key Components

1. **Models** (`pkg/models`): Domain entities and data structures
2. **Config** (`internal/config`): Configuration management with Viper
3. **Cloud-Init** (`internal/cloudinit`): Cloud-init seed generation
4. **QEMU** (`internal/qemu`): QEMU command building and execution
5. **VM Manager** (`internal/vm`): High-level VM lifecycle operations
6. **Network** (`internal/network`): Bridge networking with socket_vmnet

### Building

```bash
# Build for current platform
go build -o vm ./cmd/vm

# Build for macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o vm-arm64 ./cmd/vm

# Build for macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o vm-amd64 ./cmd/vm

# Run tests
go test ./...

# Run with race detector
go run -race ./cmd/vm create
```

### Adding New Features

1. **New VM Image**: Add to `pkg/models/config.go` in the `DefaultConfig()` function
2. **New Command**: Create a new file in `cmd/vm/commands/` and register in `root.go`
3. **New Template**: Add to `templates/` and update `internal/cloudinit/cloudinit.go`

## Troubleshooting

### QEMU Not Found

```bash
# Install manually
brew install qemu

# Or let the tool install it automatically when you run a command
```

### Bridge Networking Not Working

```bash
# Run diagnostics
vm diagnose

# Reinstall socket_vmnet
brew reinstall socket_vmnet
sudo brew services restart socket_vmnet

# Check configuration
cat /opt/homebrew/etc/socket_vmnet/config
```

### VM Won't Start

1. Check if QEMU is installed: `which qemu-system-aarch64`
2. Check VM logs: `~/.vm/vms/<vm-name>/console.log`
3. Run diagnostics: `vm diagnose`
4. Try port forwarding mode instead: `--net-type portfwd`

### Permission Denied

Make sure you're not running the command with `sudo`. The tool will ask for your password when needed.

## Comparison with Bash Version

### Advantages of Go Version

✅ **Better Performance**: Compiled binary, faster execution
✅ **Type Safety**: Compile-time error checking
✅ **Better Error Handling**: Structured error handling with context
✅ **Easier Maintenance**: Clear module boundaries and interfaces
✅ **Better Testing**: Unit testing with Go's testing framework
✅ **Cross-Platform**: Easier to port to other platforms
✅ **Dependency Management**: Go modules for version control
✅ **Professional Structure**: Industry-standard project layout

### Migration Notes

The Go version maintains compatibility with VMs created by the bash version:
- Same directory structure (`~/.vm/`)
- Same `info.json` format
- Same cloud-init configuration
- Same QEMU commands

You can use both versions interchangeably.

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

See the main project LICENSE file.

## Credits

- Original bash version: LazyLinux VM Manager
- Rewritten in Go for better maintainability and performance
- Uses [Cobra](https://github.com/spf13/cobra) for CLI
- Uses [Viper](https://github.com/spf13/viper) for configuration

## Support

For issues and questions:
- GitHub Issues: https://github.com/antoniopicone/lazylinux/issues
- Check diagnostics: `vm diagnose`
