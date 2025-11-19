package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vm",
	Short: "LazyLinux VM Manager",
	Long: `A simple and powerful VM manager for macOS using QEMU.
Complete documentation is available at https://github.com/antoniopicone/lazylinux`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
