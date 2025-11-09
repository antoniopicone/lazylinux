package qemu

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/antoniopicone/lazylinux/pkg/models"
	"github.com/antoniopicone/lazylinux/pkg/utils"
)

// Manager handles QEMU operations
type Manager struct {
	brewPrefix string
}

// NewManager creates a new QEMU manager
func NewManager(brewPrefix string) *Manager {
	return &Manager{brewPrefix: brewPrefix}
}

// BuildCommand builds the QEMU command line arguments
func (m *Manager) BuildCommand(vm *models.VM, vmDir string) ([]string, error) {
	vmDisk := filepath.Join(vmDir, "disk.qcow2")
	seedISO := filepath.Join(vmDir, "seed.iso")
	consoleLog := filepath.Join(vmDir, "console.log")
	monitorSocket := filepath.Join(vmDir, "monitor.sock")

	macAddr := utils.GenerateMACAddress(vm.Name)

	var args []string

	if vm.Architecture == models.ArchAMD64 {
		args = []string{
			"qemu-system-x86_64",
			"-machine", "q35",
			"-cpu", "qemu64",
			"-m", vm.Memory,
			"-smp", strconv.Itoa(vm.CPUs),
			"-object", fmt.Sprintf("memory-backend-ram,id=mem,size=%s", vm.Memory),
			"-numa", "node,memdev=mem",
			"-name", vm.Name,
			"-device", fmt.Sprintf("virtio-net-pci,netdev=n0,mac=%s", macAddr),
		}
	} else {
		// ARM64 configuration
		firmware, err := m.detectFirmware()
		if err != nil {
			return nil, err
		}

		args = []string{
			"qemu-system-aarch64",
			"-machine", "virt",
			"-cpu", "max",
			"-bios", firmware,
			"-m", vm.Memory,
			"-smp", strconv.Itoa(vm.CPUs),
			"-name", vm.Name,
			"-device", fmt.Sprintf("virtio-net-pci,netdev=n0,mac=%s", macAddr),
		}
	}

	// Network configuration
	if vm.NetworkType == models.NetworkBridge {
		args = append(args, "-netdev", "socket,id=n0,fd=3")
	} else {
		args = append(args, "-netdev", fmt.Sprintf("user,id=n0,hostfwd=tcp:127.0.0.1:%d-:22", vm.SSH.Port))
	}

	// Common args
	args = append(args,
		"-device", "virtio-rng-pci",
		"-drive", fmt.Sprintf("file=%s,if=virtio,cache=writeback,format=qcow2", vmDisk),
		"-cdrom", seedISO,
		"-daemonize",
		"-display", "none",
		"-serial", fmt.Sprintf("file:%s", consoleLog),
		"-monitor", fmt.Sprintf("unix:%s,server,nowait", monitorSocket),
	)

	// Add hardware acceleration for ARM64 on macOS
	if vm.Architecture == models.ArchARM64 {
		args = append(args, "-accel", "hvf")
	}

	return args, nil
}

// CreateDisk creates a VM disk from a base image
func (m *Manager) CreateDisk(baseImage, vmDisk, size string) error {
	// Convert base image to VM disk
	cmd := exec.Command("qemu-img", "convert", "-O", "qcow2", baseImage, vmDisk)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to convert image: %w", err)
	}

	// Resize disk
	cmd = exec.Command("qemu-img", "resize", vmDisk, size)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to resize disk: %w", err)
	}

	// Verify disk
	cmd = exec.Command("qemu-img", "info", vmDisk)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to verify disk: %w", err)
	}

	return nil
}

// CheckDependencies verifies that required QEMU binaries are available
func (m *Manager) CheckDependencies() error {
	requiredBins := []string{
		"qemu-system-aarch64",
		"qemu-system-x86_64",
		"qemu-img",
	}

	var missing []string
	for _, bin := range requiredBins {
		if _, err := exec.LookPath(bin); err != nil {
			missing = append(missing, bin)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required binaries: %s", strings.Join(missing, ", "))
	}

	return nil
}

// InstallQEMU installs QEMU using Homebrew
func (m *Manager) InstallQEMU() error {
	fmt.Print("Installing QEMU...")

	cmd := exec.Command("brew", "install", "qemu")
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install QEMU: %w", err)
	}

	fmt.Println(" ✓")
	return nil
}

// detectFirmware finds the ARM64 UEFI firmware
func (m *Manager) detectFirmware() (string, error) {
	candidates := []string{
		filepath.Join(m.brewPrefix, "share/qemu/edk2-aarch64-code.fd"),
		"/usr/local/share/qemu/edk2-aarch64-code.fd",
		"/usr/share/qemu/edk2-aarch64-code.fd",
		"/usr/share/edk2/aarch64/QEMU_EFI.fd",
		"/usr/share/AAVMF/AAVMF_CODE.fd",
	}

	for _, firmware := range candidates {
		if _, err := os.Stat(firmware); err == nil {
			return firmware, nil
		}
	}

	return "", fmt.Errorf("ARM64 UEFI firmware not found")
}

// IsRunning checks if a VM is currently running
func (m *Manager) IsRunning(vmName string) (bool, int, error) {
	cmd := exec.Command("pgrep", "-f", fmt.Sprintf("qemu.*-name %s", vmName))
	output, err := cmd.Output()
	if err != nil {
		// pgrep returns non-zero if no process found
		return false, 0, nil
	}

	pidStr := strings.TrimSpace(string(output))
	if pidStr == "" {
		return false, 0, nil
	}

	pid, err := strconv.Atoi(strings.Split(pidStr, "\n")[0])
	if err != nil {
		return false, 0, fmt.Errorf("failed to parse PID: %w", err)
	}

	return true, pid, nil
}

// StopVM stops a running VM gracefully
func (m *Manager) StopVM(vmName string, vmDir string) error {
	monitorSocket := filepath.Join(vmDir, "monitor.sock")

	// Try graceful shutdown via monitor
	if _, err := os.Stat(monitorSocket); err == nil {
		cmd := exec.Command("socat", "-", fmt.Sprintf("UNIX-CONNECT:%s", monitorSocket))
		cmd.Stdin = strings.NewReader("system_powerdown\nquit\n")
		_ = cmd.Run() // Ignore errors, we'll force kill if needed
	}

	// Wait a bit for graceful shutdown
	// In production, we'd wait and poll

	// Force kill if still running
	running, pid, err := m.IsRunning(vmName)
	if err != nil {
		return err
	}

	if running {
		cmd := exec.Command("kill", "-KILL", strconv.Itoa(pid))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to kill VM process: %w", err)
		}
	}

	return nil
}
