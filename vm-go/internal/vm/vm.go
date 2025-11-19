package vm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type SSHConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type Info struct {
	Name     string    `json:"name"`
	Arch     string    `json:"arch"`
	Image    string    `json:"image"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	NetType  string    `json:"net_type"`
	VirtType string    `json:"virt_type"`
	StaticIP string    `json:"static_ip,omitempty"`
	SSH      SSHConfig `json:"ssh"`
}

func SaveInfo(vmDir string, info Info) error {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(vmDir, "info.json"), data, 0644)
}

func LoadInfo(vmDir string) (*Info, error) {
	data, err := os.ReadFile(filepath.Join(vmDir, "info.json"))
	if err != nil {
		return nil, err
	}
	var info Info
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func BuildQEMUArgs(info Info, memory, cpus, diskPath, firmware string) []string {
	args := []string{}
	macAddr := generateMACAddress(info.Name)

	if info.Arch == "amd64" || info.Arch == "x86_64" {
		args = append(args, "qemu-system-x86_64", "-machine", "q35", "-cpu", "qemu64")
		args = append(args, "-m", memory, "-smp", cpus)
		args = append(args, "-object", fmt.Sprintf("memory-backend-ram,id=mem,size=%s", memory))
		args = append(args, "-numa", "node,memdev=mem")
	} else {
		args = append(args, "qemu-system-aarch64", "-machine", "virt", "-cpu", "max")
		if firmware != "" {
			args = append(args, "-bios", firmware)
		}
		args = append(args, "-m", memory, "-smp", cpus)
	}

	args = append(args, "-name", info.Name)
	args = append(args, "-device", fmt.Sprintf("virtio-net-pci,netdev=n0,mac=%s", macAddr))

	if info.NetType == "bridge" {
		args = append(args, "-netdev", "socket,id=n0,fd=3")
	} else {
		args = append(args, "-netdev", fmt.Sprintf("user,id=n0,hostfwd=tcp:127.0.0.1:%d-:22", info.SSH.Port))
	}

	args = append(args, "-device", "virtio-rng-pci")
	args = append(args, "-drive", fmt.Sprintf("file=%s,if=virtio,cache=writeback,format=qcow2", diskPath))

	if info.VirtType == "hvf" && runtime.GOOS == "darwin" {
		args = append(args, "-accel", "hvf")
	}

	return args
}

func generateMACAddress(name string) string {
	// Simple deterministic MAC generation (simplified from bash script)
	// Real implementation should match the bash script's logic if strict compatibility is needed
	// For now, using a placeholder or simple hash
	return "52:54:00:12:34:56" // TODO: Implement proper hash based MAC
}

func DetectFirmware(arch string) string {
	if arch == "amd64" {
		return ""
	}
	candidates := []string{
		"/opt/homebrew/share/qemu/edk2-aarch64-code.fd",
		"/usr/local/share/qemu/edk2-aarch64-code.fd",
		"/usr/share/qemu/edk2-aarch64-code.fd",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}
