package cmd

import (
	"fmt"
	"lazylinux-vm/internal/config"
	"lazylinux-vm/internal/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [vm-name]",
	Short: "Stop a running VM",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		vmDir := filepath.Join(config.VMsDir, vmName)

		if _, err := os.Stat(vmDir); os.IsNotExist(err) {
			utils.Error("VM '%s' does not exist", vmName)
		}

		utils.Log("Stopping VM: %s", vmName)

		pidFile := filepath.Join(vmDir, "qemu.pid")
		pidData, err := os.ReadFile(pidFile)
		if err != nil {
			utils.Log("VM '%s' is not running", vmName)
			return
		}

		pid, err := strconv.Atoi(string(pidData))
		if err != nil || pid <= 0 {
			utils.Log("Invalid PID file")
			return
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			utils.Log("VM '%s' is not running", vmName)
			return
		}

		// Try graceful shutdown via monitor socket
		monitorSocket := filepath.Join(vmDir, "monitor.sock")
		if _, err := os.Stat(monitorSocket); err == nil {
			utils.Log("Attempting graceful shutdown...")
			cmd := exec.Command("sh", "-c", fmt.Sprintf("echo 'system_powerdown' | nc -U %s", monitorSocket))
			cmd.Run()
		}

		// Wait for process to exit
		utils.Log("Waiting for VM to shut down...")
		for i := 0; i < 30; i++ {
			if err := process.Signal(syscall.Signal(0)); err != nil {
				utils.Log("VM stopped successfully")
				os.Remove(pidFile)
				return
			}
			time.Sleep(1 * time.Second)
		}

		// Force kill if still running
		utils.Log("Graceful shutdown timed out, forcing stop...")
		process.Kill()
		time.Sleep(2 * time.Second)
		os.Remove(pidFile)
		utils.Log("VM stopped successfully")
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
