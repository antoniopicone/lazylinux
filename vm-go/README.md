# VM-Go - LazyLinux VM Manager (Go Implementation)

This is a Go port of the original Bash `vm` script, providing a more robust and maintainable VM management solution for macOS using QEMU.

## Features

- ✅ Create and manage QEMU VMs on macOS
- ✅ Support for both ARM64 and AMD64 architectures
- ✅ Bridge networking and port forwarding modes
- ✅ Cloud-init based VM initialization
- ✅ Automatic IP assignment
- ✅ HVF acceleration support for ARM64 VMs

## Project Structure

```
vm-go/
├── main.go              # Entry point
├── cmd/                 # CLI commands
│   ├── root.go         # Root command
│   ├── create.go       # Create VM
│   ├── list.go         # List VMs
│   ├── stop.go         # Stop VM
│   └── delete.go       # Delete VM
├── internal/
│   ├── config/         # Configuration
│   ├── vm/             # VM management
│   ├── network/        # Networking utilities
│   ├── cloudinit/      # Cloud-init generation
│   ├── image/          # Image management
│   └── utils/          # Helper functions
└── test-suite.sh       # Comprehensive test suite
```

## Installation

### Prerequisites

- Go 1.21 or later
- QEMU (`brew install qemu`)
- socket_vmnet for bridge networking (`brew install socket_vmnet`)

### Build

```bash
cd vm-go
go build -o vm-go
```

## Usage

### Create a VM

```bash
# Create with default settings (random name, bridge networking)
./vm-go create

# Create with specific name and settings
./vm-go create --name myvm --arch arm64 --memory 2G --cpus 2

# Create with port forwarding (no bridge setup needed)
./vm-go create --name myvm --net-type portfwd
```

### List VMs

```bash
./vm-go list
```

### Stop a VM

```bash
./vm-go stop myvm
```

### Delete a VM

```bash
./vm-go delete myvm

# Force delete without confirmation
./vm-go delete myvm --force
```

## Testing

A comprehensive test suite is included to verify all functionality:

```bash
./test-suite.sh
```

The test suite will:
1. ✅ Verify binary exists and is executable
2. ✅ Test help and list commands
3. ✅ Create multiple test VMs
4. ✅ Verify VMs appear in list
5. ✅ Check VM running status
6. ✅ Extract SSH connection info
7. ✅ Wait for VMs to boot
8. ✅ Test SSH connectivity
9. ✅ Test stop command
10. ✅ Clean up and delete test VMs
11. ✅ Verify deletion

A detailed report will be generated in `test-report-YYYYMMDD-HHMMSS.txt`.

## Optimizations vs Original Bash Script

### Performance
- **Faster execution**: No shell overhead, compiled binary
- **Better concurrency**: Can easily add parallel VM creation
- **Efficient JSON parsing**: Native JSON support vs grep/sed

### Code Quality
- **Type safety**: Compile-time type checking
- **Better error handling**: Structured error propagation
- **Maintainability**: Clear package structure and separation of concerns

### Features
- **Structured configuration**: Proper structs instead of string parsing
- **Better testing**: Unit tests for individual components
- **Cross-platform potential**: Easier to port to other platforms

## Differences from Original Script

### Implemented
- ✅ VM creation with cloud-init
- ✅ VM listing with status
- ✅ VM stopping (graceful shutdown)
- ✅ VM deletion
- ✅ Network configuration (bridge and portfwd)
- ✅ Image downloading
- ✅ HVF acceleration support

### Not Yet Implemented
- ⏳ `start` command (restart stopped VM)
- ⏳ `ssh` command (direct SSH into VM)
- ⏳ `ip` command (get VM IP address)
- ⏳ `gui` command (start with graphical display)
- ⏳ `install` command (system installation)
- ⏳ `setup-bridge` command
- ⏳ `diagnose` command
- ⏳ Multi-VM creation (comma-separated names)
- ⏳ Progress indicators and spinners

## Development

### Adding a New Command

1. Create a new file in `cmd/` (e.g., `cmd/mycommand.go`)
2. Define the command using Cobra
3. Add the command to `rootCmd` in the `init()` function
4. Implement the command logic

Example:
```go
package cmd

import (
    "github.com/spf13/cobra"
)

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Description",
    Run: func(cmd *cobra.Command, args []string) {
        // Implementation
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
}
```

### Running Tests

```bash
# Run the test suite
./test-suite.sh

# Build and test
go build -o vm-go && ./test-suite.sh
```

## License

Same as the original LazyLinux project.

## Contributing

Contributions are welcome! Please ensure:
- Code follows Go conventions
- Tests pass
- Documentation is updated
