package config

import (
	"os"
	"path/filepath"
)

const (
	DefaultUsername = "user01"
	DefaultMemory   = "2G"
	DefaultCPUs     = "2"
	DefaultDisk     = "10G"
	DefaultImage    = "debian13"
)

var (
	WorkRoot   string
	ImagesDir  string
	VMsDir     string
	BrewPrefix string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	WorkRoot = filepath.Join(home, ".vm")
	ImagesDir = filepath.Join(WorkRoot, "images")
	VMsDir = filepath.Join(WorkRoot, "vms")

	// Simple check for Homebrew prefix
	if _, err := os.Stat("/opt/homebrew"); err == nil {
		BrewPrefix = "/opt/homebrew"
	} else {
		BrewPrefix = "/usr/local"
	}
}

func GetSocketVMNetSocket() string {
	return filepath.Join(BrewPrefix, "var/run/socket_vmnet")
}

func GetSocketVMNetClient() string {
	return filepath.Join(BrewPrefix, "opt/socket_vmnet/bin/socket_vmnet_client")
}
