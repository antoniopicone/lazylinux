package cmd

import (
	"fmt"
	"lazylinux-vm/internal/config"
	"lazylinux-vm/internal/utils"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [vm-name]",
	Short: "Delete a VM and its files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		vmDir := filepath.Join(config.VMsDir, vmName)

		if _, err := os.Stat(vmDir); os.IsNotExist(err) {
			utils.Error("VM '%s' does not exist", vmName)
		}

		// Check if VM is running and stop it
		pidFile := filepath.Join(vmDir, "qemu.pid")
		if _, err := os.Stat(pidFile); err == nil {
			utils.Log("Stopping running VM before deletion...")
			stopCmd.Run(cmd, args)
		}

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("This will permanently delete VM '%s' and all its data.\n", vmName)
			fmt.Print("Are you sure? (y/N): ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				utils.Log("Deletion cancelled")
				return
			}
		}

		utils.Log("Deleting VM: %s", vmName)
		if err := os.RemoveAll(vmDir); err != nil {
			utils.Error("Failed to delete VM: %v", err)
		}
		utils.Log("VM '%s' deleted successfully", vmName)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().Bool("force", false, "Delete without confirmation")
}
