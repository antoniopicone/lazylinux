package cmd

import (
	"fmt"
	"lazylinux-vm/internal/config"
	"lazylinux-vm/internal/vm"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all VMs and their status",
	Run: func(cmd *cobra.Command, args []string) {
		if err := os.MkdirAll(config.VMsDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating VMs directory: %v\n", err)
			return
		}

		entries, err := os.ReadDir(config.VMsDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading VMs directory: %v\n", err)
			return
		}

		if len(entries) == 0 {
			fmt.Println("No VMs found")
			return
		}

		fmt.Println()
		fmt.Printf("%-20s %-10s %-8s %-10s %-25s %-s\n", "NAME", "STATUS", "ARCH", "IMAGE", "IP ADDRESS", "CREDENTIALS")
		fmt.Printf("%-20s %-10s %-8s %-10s %-25s %-s\n", "----", "------", "----", "-----", "----------", "-----------")

		vmCount := 0
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			vmName := entry.Name()
			vmDir := filepath.Join(config.VMsDir, vmName)
			info, err := vm.LoadInfo(vmDir)

			if err != nil {
				fmt.Printf("%-20s %-10s %-8s %-10s %-25s %-s\n", vmName, "BROKEN", "-", "-", "-", "-")
				vmCount++
				continue
			}

			status := "STOPPED"
			sshInfo := "-"
			credentials := "-"

			// Check if VM is running
			pidFile := filepath.Join(vmDir, "qemu.pid")
			if pidData, err := os.ReadFile(pidFile); err == nil {
				pid, _ := strconv.Atoi(string(pidData))
				if pid > 0 {
					// Check if process exists
					process, err := os.FindProcess(pid)
					if err == nil {
						err = process.Signal(syscall.Signal(0))
						if err == nil {
							status = "RUNNING"
							if info.NetType == "bridge" {
								if info.SSH.Host != "" && info.SSH.Host != "dhcp-assigned" {
									sshInfo = info.SSH.Host
								} else {
									sshInfo = fmt.Sprintf("%s.local", vmName)
								}
							} else {
								sshInfo = fmt.Sprintf("127.0.0.1:%d", info.SSH.Port)
							}
							credentials = fmt.Sprintf("%s / %s", info.Username, info.Password)
						}
					}
				}
			}

			arch := info.Arch
			if arch == "" {
				arch = "arm64"
			}
			image := info.Image
			if image == "" {
				image = "unknown"
			}

			fmt.Printf("%-20s %-10s %-8s %-10s %-25s %-s\n", vmName, status, arch, image, sshInfo, credentials)
			vmCount++
		}

		fmt.Println()
		fmt.Printf("Total VMs: %d\n", vmCount)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
