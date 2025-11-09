# Quick Start Guide

## Build & Install

```bash
# Build the binary
make build
# or
go build -o vm ./cmd/vm

# Install system-wide
make install
# or
sudo cp vm /usr/local/bin/
```

## First VM

```bash
# Setup bridge networking (one-time)
vm setup-bridge

# Create your first VM
vm create

# List VMs
vm list

# SSH into VM
vm ssh <vm-name>
```

## Common Commands

```bash
# Create VM with specific settings
vm create --name myvm --memory 8G --cpus 4

# Create multiple VMs at once
vm create --name vm1,vm2,vm3 --user admin --pass test

# Create VM with port forwarding (no bridge setup needed)
vm create --name dev --net-type portfwd

# Start/Stop VMs
vm start myvm
vm stop myvm

# Delete VM
vm delete myvm --force

# Get VM IP
vm ip myvm

# Cleanup orphaned processes
vm cleanup

# Delete all VMs
vm purge --force

# Diagnostics
vm diagnose
```

## Troubleshooting

```bash
# Check if everything is OK
vm diagnose

# If QEMU is missing, it will auto-install when you run a command

# If bridge networking fails, try port forwarding
vm create --net-type portfwd
```

## Architecture

The project is organized as follows:

- **cmd/vm** - CLI entry point
- **internal/** - Private application code
  - **cloudinit/** - Cloud-init seed generation
  - **config/** - Configuration management
  - **network/** - Bridge networking setup
  - **qemu/** - QEMU operations
  - **vm/** - VM lifecycle management
- **pkg/** - Public libraries
  - **models/** - Domain models
  - **utils/** - Utilities
- **templates/** - Cloud-init templates (also embedded in code)

## Development

```bash
# Run tests
make test

# Format code
make fmt

# Run linter
make lint

# Clean build artifacts
make clean
```

## Binary Size

The compiled binary is ~14MB for ARM64. This includes all dependencies and templates embedded.

To reduce size:
```bash
# Build with optimizations
go build -ldflags="-s -w" -o vm ./cmd/vm

# Further compress with UPX (optional)
upx --best --lzma vm
```

## Notes

- The Go version is fully compatible with VMs created by the bash version
- Both versions can coexist and manage the same VMs
- Configuration is optional (uses sensible defaults)
- Templates are embedded in the binary (no external files needed at runtime)
