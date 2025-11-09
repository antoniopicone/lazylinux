package utils

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"regexp"
)

// ValidateIP checks if the given string is a valid IPv4 address
func ValidateIP(ip string) bool {
	return net.ParseIP(ip) != nil && net.ParseIP(ip).To4() != nil
}

// GenerateStaticIP generates a deterministic IP address based on VM name
// Returns an IP in the range 192.168.105.100-254
func GenerateStaticIP(vmName string) string {
	hash := sha256.Sum256([]byte(vmName))
	// Use first 4 bytes of hash to generate a number
	num := binary.BigEndian.Uint32(hash[:4])
	// Map to range 100-254 (155 possible values)
	lastOctet := 100 + (num % 155)
	return fmt.Sprintf("192.168.105.%d", lastOctet)
}

// GenerateMACAddress generates a deterministic MAC address based on VM name
// Uses QEMU's default prefix 52:54:00
func GenerateMACAddress(vmName string) string {
	hash := sha256.Sum256([]byte(vmName))
	return fmt.Sprintf("52:54:00:%02x:%02x:%02x", hash[0], hash[1], hash[2])
}

// SanitizeHostname replaces underscores with hyphens for RFC 952/1123 compliance
func SanitizeHostname(name string) string {
	re := regexp.MustCompile(`_`)
	return re.ReplaceAllString(name, "-")
}

// IsPortFree checks if a TCP port is available
func IsPortFree(port int) bool {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// FindFreePort finds an available TCP port starting from preferredPort
func FindFreePort(preferredPort int) (int, error) {
	if preferredPort > 0 && IsPortFree(preferredPort) {
		return preferredPort, nil
	}

	// Skip commonly used ports
	skipPorts := map[int]bool{
		3000: true, 3306: true, 5000: true, 5432: true,
		8000: true, 8080: true, 8443: true, 9000: true,
	}

	for port := 2222; port <= 9999; port++ {
		if skipPorts[port] {
			continue
		}
		if IsPortFree(port) {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no free port found in range 2222-9999")
}
