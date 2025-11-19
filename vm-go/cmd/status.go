package cmd

import (
	"bufio"
	"fmt"
	"lazylinux-vm/internal/config"
	"lazylinux-vm/internal/utils"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [vm-name]",
	Short: "Check VM boot status and cloud-init progress",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		vmDir := filepath.Join(config.VMsDir, vmName)

		if _, err := os.Stat(vmDir); os.IsNotExist(err) {
			utils.Error("VM '%s' does not exist", vmName)
		}

		consoleLog := filepath.Join(vmDir, "console.log")
		if _, err := os.Stat(consoleLog); os.IsNotExist(err) {
			fmt.Println("VM has not started yet (no console log)")
			return
		}

		// Check for cloud-init completion
		file, err := os.Open(consoleLog)
		if err != nil {
			utils.Error("Failed to read console log: %v", err)
		}
		defer file.Close()

		cloudInitReady := false
		hasNetworkInfo := false
		ipAddress := ""

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()

			if strings.Contains(line, "CLOUD-INIT-READY") {
				cloudInitReady = true
			}

			// Look for network info
			if strings.Contains(line, "ci-info") && strings.Contains(line, "enp0s1") {
				hasNetworkInfo = true
				// Try to extract IP
				fields := strings.Fields(line)
				for _, field := range fields {
					if strings.Contains(field, "192.168.") || strings.Contains(field, "10.") {
						ipAddress = field
					}
				}
			}
		}

		fmt.Printf("VM: %s\n", vmName)
		fmt.Printf("Console log: %s\n", consoleLog)
		fmt.Println()

		if cloudInitReady {
			fmt.Printf("%s✔%s Cloud-init: READY\n", utils.Green, utils.NC)
		} else {
			fmt.Printf("%s⏳%s Cloud-init: Still initializing...\n", utils.Yellow, utils.NC)
		}

		if hasNetworkInfo {
			fmt.Printf("%s✔%s Network: Configured\n", utils.Green, utils.NC)
			if ipAddress != "" {
				fmt.Printf("   IP Address: %s\n", ipAddress)
			}
		} else {
			fmt.Printf("%s⏳%s Network: Not yet configured\n", utils.Yellow, utils.NC)
		}

		fmt.Println()
		fmt.Println("To view full console output:")
		fmt.Printf("  tail -f %s\n", consoleLog)
	},
}

var waitCmd = &cobra.Command{
	Use:   "wait [vm-name]",
	Short: "Wait for VM to complete boot and cloud-init",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		vmDir := filepath.Join(config.VMsDir, vmName)

		if _, err := os.Stat(vmDir); os.IsNotExist(err) {
			utils.Error("VM '%s' does not exist", vmName)
		}

		consoleLog := filepath.Join(vmDir, "console.log")
		timeout, _ := cmd.Flags().GetInt("timeout")

		fmt.Printf("Waiting for VM '%s' to complete boot (timeout: %ds)...\n", vmName, timeout)

		start := time.Now()
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			if time.Since(start).Seconds() > float64(timeout) {
				fmt.Printf("%s✗%s Timeout waiting for VM to boot\n", utils.Red, utils.NC)
				os.Exit(1)
			}

			if _, err := os.Stat(consoleLog); err == nil {
				file, err := os.Open(consoleLog)
				if err == nil {
					scanner := bufio.NewScanner(file)
					for scanner.Scan() {
						if strings.Contains(scanner.Text(), "CLOUD-INIT-READY") {
							file.Close()
							fmt.Printf("%s✔%s VM is ready!\n", utils.Green, utils.NC)

							// Try to extract IP
							file2, _ := os.Open(consoleLog)
							scanner2 := bufio.NewScanner(file2)
							for scanner2.Scan() {
								line := scanner2.Text()
								if strings.Contains(line, "ci-info") && strings.Contains(line, "enp0s1") {
									fields := strings.Fields(line)
									for _, field := range fields {
										if strings.Contains(field, "192.168.") || strings.Contains(field, "10.") {
											fmt.Printf("IP Address: %s\n", field)
											break
										}
									}
								}
							}
							file2.Close()
							return
						}
					}
					file.Close()
				}
			}

			fmt.Print(".")
			<-ticker.C
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(waitCmd)

	waitCmd.Flags().Int("timeout", 300, "Timeout in seconds")
}
