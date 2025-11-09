package models

// Config represents the global configuration for LazyLinux
type Config struct {
	WorkRoot         string       `yaml:"work_root"`
	ImagesDir        string       `yaml:"images_dir"`
	VMsDir           string       `yaml:"vms_dir"`
	DefaultUsername  string       `yaml:"default_username"`
	DefaultMemory    string       `yaml:"default_memory"`
	DefaultCPUs      int          `yaml:"default_cpus"`
	DefaultDiskSize  string       `yaml:"default_disk_size"`
	DefaultImage     string       `yaml:"default_image"`
	BrewPrefix       string       `yaml:"brew_prefix"`
	SocketVMNetPath  string       `yaml:"socket_vmnet_path"`
	Images           []ImageInfo  `yaml:"images"`
}

// ImageInfo contains information about available VM images
type ImageInfo struct {
	Name        string            `yaml:"name"`
	DisplayName string            `yaml:"display_name"`
	URLs        map[string]string `yaml:"urls"` // arch -> URL
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		WorkRoot:        "$HOME/.vm",
		ImagesDir:       "$HOME/.vm/images",
		VMsDir:          "$HOME/.vm/vms",
		DefaultUsername: "user01",
		DefaultMemory:   "4G",
		DefaultCPUs:     2,
		DefaultDiskSize: "10G",
		DefaultImage:    "debian13",
		BrewPrefix:      "/opt/homebrew",
		Images: []ImageInfo{
			{
				Name:        "debian13",
				DisplayName: "Debian 13 (Trixie)",
				URLs: map[string]string{
					"arm64": "https://cloud.debian.org/images/cloud/trixie/latest/debian-13-genericcloud-arm64.qcow2",
					"amd64": "https://cloud.debian.org/images/cloud/trixie/latest/debian-13-genericcloud-amd64.qcow2",
				},
			},
			{
				Name:        "debian12",
				DisplayName: "Debian 12 (Bookworm)",
				URLs: map[string]string{
					"arm64": "https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-genericcloud-arm64.qcow2",
					"amd64": "https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-generic-amd64.qcow2",
				},
			},
		},
	}
}
