package models

import "time"

// Architecture represents the CPU architecture of a VM
type Architecture string

const (
	ArchARM64 Architecture = "arm64"
	ArchAMD64 Architecture = "amd64"
)

// NetworkType represents the type of network configuration
type NetworkType string

const (
	NetworkBridge  NetworkType = "bridge"
	NetworkPortFwd NetworkType = "portfwd"
)

// VMStatus represents the current state of a VM
type VMStatus string

const (
	StatusRunning VMStatus = "running"
	StatusStopped VMStatus = "stopped"
	StatusBroken  VMStatus = "broken"
)

// VM represents a virtual machine configuration
type VM struct {
	Name         string       `json:"name"`
	Architecture Architecture `json:"arch"`
	Image        string       `json:"image"`
	Username     string       `json:"username"`
	Password     string       `json:"password"`
	NetworkType  NetworkType  `json:"net_type"`
	StaticIP     string       `json:"static_ip,omitempty"`
	SSH          SSHConfig    `json:"ssh"`
	Memory       string       `json:"memory,omitempty"`
	CPUs         int          `json:"cpus,omitempty"`
	DiskSize     string       `json:"disk_size,omitempty"`
	CreatedAt    time.Time    `json:"created_at,omitempty"`
}

// SSHConfig contains SSH connection information
type SSHConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// VMCreateOptions contains options for creating a new VM
type VMCreateOptions struct {
	Name         string
	Architecture Architecture
	Username     string
	Password     string
	Memory       string
	CPUs         int
	DiskSize     string
	SSHPort      int
	NetworkType  NetworkType
	StaticIP     string
	ShowConsole  bool
	Debug        bool
}

// VMInfo contains runtime information about a VM
type VMInfo struct {
	VM       *VM
	Status   VMStatus
	PID      int
	DiskPath string
	LogPath  string
}
