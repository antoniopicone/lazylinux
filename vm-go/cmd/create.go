package cmd

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"lazylinux-vm/internal/cloudinit"
	"lazylinux-vm/internal/config"
	"lazylinux-vm/internal/image"
	"lazylinux-vm/internal/network"
	"lazylinux-vm/internal/utils"
	"lazylinux-vm/internal/vm"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	vmName   string
	arch     string
	username string
	password string
	memory   string
	cpus     string
	disk     string
	netType  string
	virtType string
	staticIP string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create and start a new VM",
	Run: func(cmd *cobra.Command, args []string) {
		utils.CheckRoot()

		// Generate random name if not provided
		if vmName == "" {
			vmName = generateRandomVMName()
		}

		// Sanitize hostname
		vmName = network.SanitizeHostname(vmName)

		// Set default virt type based on arch
		if virtType == "" {
			if arch == "arm64" {
				virtType = "hvf"
			} else {
				virtType = "qemu"
			}
		}

		// Generate password if not provided
		if password == "" {
			password = generatePassword(16)
		}

		// Generate static IP for bridge mode
		if netType == "bridge" && staticIP == "" {
			staticIP = network.GenerateStaticIP(vmName)
		}

		vmDir := filepath.Join(config.VMsDir, vmName)
		if _, err := os.Stat(vmDir); err == nil {
			utils.Error("VM '%s' already exists", vmName)
		}

		if err := os.MkdirAll(vmDir, 0755); err != nil {
			utils.Error("Failed to create VM directory: %v", err)
		}

		utils.Log("Creating VM: %s", vmName)

		// Download base image
		baseImage, err := image.DownloadImage(config.DefaultImage, arch)
		if err != nil {
			utils.Error("Failed to download image: %v", err)
		}

		// Create VM disk
		vmDisk := filepath.Join(vmDir, "disk.qcow2")
		qemuConvertCmd := exec.Command("qemu-img", "convert", "-O", "qcow2", baseImage, vmDisk)
		if err := qemuConvertCmd.Run(); err != nil {
			utils.Error("Failed to create VM disk: %v", err)
		}

		qemuResizeCmd := exec.Command("qemu-img", "resize", vmDisk, disk)
		if err := qemuResizeCmd.Run(); err != nil {
			utils.Error("Failed to resize VM disk: %v", err)
		}

		// Create cloud-init ISO
		ciConfig := cloudinit.Config{
			Hostname: vmName,
			Username: username,
			Password: password,
			NetType:  netType,
			StaticIP: staticIP,
		}

		seedISO, err := cloudinit.CreateISO(vmDir, ciConfig)
		if err != nil {
			utils.Error("Failed to create cloud-init ISO: %v", err)
		}

		// Find free port for portfwd mode
		sshPort := 22
		if netType == "portfwd" {
			sshPort, err = network.FindFreePort()
			if err != nil {
				utils.Error("Failed to find free port: %v", err)
			}
		}

		// Save VM info
		info := vm.Info{
			Name:     vmName,
			Arch:     arch,
			Image:    config.DefaultImage,
			Username: username,
			Password: password,
			NetType:  netType,
			VirtType: virtType,
			StaticIP: staticIP,
			SSH: vm.SSHConfig{
				Host: staticIP,
				Port: sshPort,
			},
		}

		if netType == "portfwd" {
			info.SSH.Host = "127.0.0.1"
		}

		if err := vm.SaveInfo(vmDir, info); err != nil {
			utils.Error("Failed to save VM info: %v", err)
		}

		// Build QEMU command
		firmware := vm.DetectFirmware(arch)
		qemuArgs := vm.BuildQEMUArgs(info, memory, cpus, vmDisk, firmware)

		// Add cloud-init ISO
		qemuArgs = append(qemuArgs, "-cdrom", seedISO)

		// Add daemon options
		consoleLog := filepath.Join(vmDir, "console.log")
		pidFile := filepath.Join(vmDir, "qemu.pid")
		monitorSocket := filepath.Join(vmDir, "monitor.sock")

		qemuArgs = append(qemuArgs,
			"-daemonize",
			"-display", "none",
			"-serial", fmt.Sprintf("file:%s", consoleLog),
			"-monitor", fmt.Sprintf("unix:%s,server,nowait", monitorSocket),
			"-pidfile", pidFile,
		)

		// Start QEMU
		utils.Log("Starting VM...")

		var qemuCmd *exec.Cmd
		if netType == "bridge" {
			// For bridge mode, use socket_vmnet_client
			socketVMNetClient := config.GetSocketVMNetClient()
			socketVMNetSocket := config.GetSocketVMNetSocket()

			allArgs := append([]string{socketVMNetSocket}, qemuArgs...)
			qemuCmd = exec.Command("sudo", append([]string{socketVMNetClient}, allArgs...)...)
		} else {
			qemuCmd = exec.Command(qemuArgs[0], qemuArgs[1:]...)
		}

		if err := qemuCmd.Run(); err != nil {
			utils.Error("Failed to start QEMU: %v", err)
		}

		// Wait for cloud-init to complete
		utils.Log("Waiting for VM to complete boot (this may take 1-2 minutes)...")
		if err := waitForCloudInit(consoleLog, 300, vmName); err != nil {
			fmt.Printf("%s⚠%s  VM started but cloud-init did not complete within timeout\n", utils.Yellow, utils.NC)
			fmt.Printf("   You can check status with: ./vm-go status %s\n", vmName)
		} else {
			// Try to extract actual IP from console log
			actualIP := extractIPFromLog(consoleLog)
			if actualIP != "" && netType == "bridge" {
				staticIP = actualIP
				// Update info.json with actual IP
				info.SSH.Host = actualIP
				vm.SaveInfo(vmDir, info)
			}
		}

		fmt.Printf("%s✔%s VM '%s' created successfully\n", utils.Green, utils.NC, vmName)
		fmt.Printf("Username: %s\n", username)
		fmt.Printf("Password: %s\n", password)
		if netType == "bridge" {
			fmt.Printf("IP: %s\n", staticIP)
			fmt.Printf("SSH: ssh %s@%s\n", username, staticIP)
		} else {
			fmt.Printf("SSH: ssh %s@127.0.0.1 -p %d\n", username, sshPort)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVar(&vmName, "name", "", "VM name")
	createCmd.Flags().StringVar(&arch, "arch", "arm64", "Architecture (arm64, amd64)")
	createCmd.Flags().StringVar(&username, "user", config.DefaultUsername, "Username")
	createCmd.Flags().StringVar(&password, "pass", "", "Password (auto-generated if not provided)")
	createCmd.Flags().StringVar(&memory, "memory", config.DefaultMemory, "Memory size")
	createCmd.Flags().StringVar(&cpus, "cpus", config.DefaultCPUs, "CPU count")
	createCmd.Flags().StringVar(&disk, "disk", config.DefaultDisk, "Disk size")
	createCmd.Flags().StringVar(&netType, "net-type", "bridge", "Network type (bridge, portfwd)")
	createCmd.Flags().StringVar(&virtType, "virt", "", "Virtualization type (qemu, hvf)")
	createCmd.Flags().StringVar(&staticIP, "ip", "", "Static IP address for bridge mode")
}

func generateRandomVMName() string {
	adjectives := []string{"swift", "bright", "clever", "happy", "quick", "bold", "calm", "deep"}
	nouns := []string{"server", "engine", "cloud", "node", "host", "box", "core", "hub"}

	adjIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(adjectives))))
	nounIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(nouns))))
	numIdx, _ := rand.Int(rand.Reader, big.NewInt(900))

	return fmt.Sprintf("%s-%s-%d", adjectives[adjIdx.Int64()], nouns[nounIdx.Int64()], 100+numIdx.Int64())
}

func generatePassword(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[idx.Int64()]
	}
	return string(b)
}

// waitForCloudInit waits for cloud-init to complete by monitoring the console log
func waitForCloudInit(consoleLog string, timeoutSeconds int, vmName string) error {
	start := time.Now()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	dots := 0
	for {
		if time.Since(start).Seconds() > float64(timeoutSeconds) {
			fmt.Println()
			return fmt.Errorf("timeout")
		}

		if _, err := os.Stat(consoleLog); err == nil {
			file, err := os.Open(consoleLog)
			if err == nil {
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					if strings.Contains(scanner.Text(), "CLOUD-INIT-READY") {
						file.Close()
						fmt.Println()
						return nil
					}
				}
				file.Close()
			}
		}

		// Show progress dots
		fmt.Print(".")
		dots++
		if dots >= 50 {
			fmt.Println()
			dots = 0
		}

		<-ticker.C
	}
}

// extractIPFromLog tries to extract the VM's IP address from the console log
func extractIPFromLog(consoleLog string) string {
	file, err := os.Open(consoleLog)
	if err != nil {
		return ""
	}
	defer file.Close()

	ipRegex := regexp.MustCompile(`\b(?:192\.168\.|10\.)\d{1,3}\.\d{1,3}\b`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "ci-info") && strings.Contains(line, "enp0s1") {
			if match := ipRegex.FindString(line); match != "" {
				return match
			}
		}
	}

	return ""
}
