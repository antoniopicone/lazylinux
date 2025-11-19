package network

import (
	"fmt"
	"hash/crc32"
	"net"
	"strings"
)

// GenerateStaticIP generates a deterministic IP based on the VM name.
// It maps the name to the range 192.168.105.100-254.
func GenerateStaticIP(vmName string) string {
	checksum := crc32.ChecksumIEEE([]byte(vmName))
	lastOctet := 100 + (checksum % 155)
	return fmt.Sprintf("192.168.105.%d", lastOctet)
}

// ValidateIP checks if the given string is a valid IPv4 address.
func ValidateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() != nil
}

// SanitizeHostname replaces underscores with hyphens for RFC compliance.
func SanitizeHostname(name string) string {
	return strings.ReplaceAll(name, "_", "-")
}

// IsPortFree checks if a TCP port is free on localhost.
func IsPortFree(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// FindFreePort finds a free port starting from 2222.
func FindFreePort() (int, error) {
	for port := 2222; port < 10000; port++ {
		// Skip common ports
		switch port {
		case 3000, 3306, 5000, 5432, 8000, 8080, 8443, 9000:
			continue
		}
		if IsPortFree(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no free port found")
}
