package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/antoniopicone/lazylinux/pkg/models"
	"github.com/spf13/viper"
)

// Manager handles configuration loading and management
type Manager struct {
	config *models.Config
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		config: models.DefaultConfig(),
	}
}

// Load loads configuration from file and environment
func (m *Manager) Load() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Look for config in ~/.vm directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".vm")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("work_root", filepath.Join(homeDir, ".vm"))
	viper.SetDefault("default_username", "user01")
	viper.SetDefault("default_memory", "4G")
	viper.SetDefault("default_cpus", 2)
	viper.SetDefault("default_disk_size", "10G")
	viper.SetDefault("default_image", "debian13")

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// Unmarshal into config struct
	if err := viper.Unmarshal(m.config); err != nil {
		return err
	}

	// Expand environment variables in paths
	m.config.WorkRoot = expandPath(m.config.WorkRoot)
	m.config.ImagesDir = expandPath(m.config.ImagesDir)
	m.config.VMsDir = expandPath(m.config.VMsDir)

	// Set derived paths if not explicitly set
	if m.config.ImagesDir == "" {
		m.config.ImagesDir = filepath.Join(m.config.WorkRoot, "images")
	}
	if m.config.VMsDir == "" {
		m.config.VMsDir = filepath.Join(m.config.WorkRoot, "vms")
	}

	// Create directories if they don't exist
	if err := os.MkdirAll(m.config.ImagesDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(m.config.VMsDir, 0755); err != nil {
		return err
	}

	return nil
}

// Get returns the current configuration
func (m *Manager) Get() *models.Config {
	return m.config
}

// expandPath expands ~ and environment variables in a path
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, _ := os.UserHomeDir()
		path = filepath.Join(homeDir, path[2:])
	}
	return os.ExpandEnv(path)
}
