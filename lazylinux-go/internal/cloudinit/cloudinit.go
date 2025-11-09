package cloudinit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/antoniopicone/lazylinux/pkg/models"
)

// bridgeTemplate is the cloud-init template for bridge networking
var bridgeTemplate = `#cloud-config
datasource_list: [ NoCloud, None ]
final_message: "CLOUD-INIT-READY"
hostname: {{.Hostname}}
fqdn: {{.Hostname}}.local
manage_etc_hosts: true
package_update: false
packages:
  - openssh-server
  - avahi-daemon
  - locales-all
output:
  all: '| tee -a /var/log/cloud-init-output.log'
write_files:
  - path: /etc/systemd/system/systemd-networkd-wait-online.service.d/override.conf
    content: |
      [Service]
      ExecStart=
      ExecStart=/lib/systemd/systemd-networkd-wait-online --timeout=3
  - path: /etc/cloud/cloud.cfg.d/99-disable-network-config.cfg
    content: |
      network: {config: disabled}
  - path: /etc/netplan/01-netcfg.yaml
    content: |
      network:
        version: 2
        renderer: networkd
        ethernets:
          enp0s1:
{{- if .StaticIP}}
            addresses:
              - {{.StaticIP}}/24
            routes:
              - to: default
                via: 192.168.105.1
            nameservers:
              addresses: [8.8.8.8, 1.1.1.1]
            dhcp4: no
            dhcp6: no
{{- else}}
            dhcp4: yes
            dhcp6: no
            nameservers:
              addresses: [8.8.8.8, 1.1.1.1]
{{- end}}
  - path: /usr/local/bin/report-ip.sh
    permissions: '0755'
    content: |
      #!/bin/bash
      sleep 5
      IP=$(ip -4 addr show enp0s1 | grep -oP '(?<=inet\s)\d+(\.\d+){3}')
      if [ -n "$IP" ]; then
          echo "VM_IP_ADDRESS: $IP"
          logger "VM_IP_ADDRESS: $IP"
      fi
  - path: /etc/ssh/sshd_config.d/50-cloud-init.conf
    owner: root:root
    permissions: '0644'
    content: |
      PasswordAuthentication yes
      PubkeyAuthentication yes
      PermitRootLogin prohibit-password
      UseDNS no
      GSSAPIAuthentication no
users:
  - name: {{.Username}}
    gecos: {{.Username}}
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: [sudo]
    shell: /bin/bash
    lock_passwd: false
    plain_text_passwd: '{{.Password}}'
ssh_pwauth: true
chpasswd:
  list: |
    {{.Username}}:{{.Password}}
  expire: false
runcmd:
  - |
    hostnamectl set-hostname {{.Hostname}} || echo "{{.Hostname}}" > /etc/hostname
    hostname {{.Hostname}}
  - echo "{{.Username}}:{{.Password}}" | chpasswd
  - passwd -u {{.Username}}
  - systemctl enable ssh
  - systemctl enable avahi-daemon
  - |
    if [ ! -f /etc/ssh/.keys_generated ]; then
      ssh-keygen -A
      touch /etc/ssh/.keys_generated
    fi
  - mkdir -p /etc/systemd/system/systemd-networkd-wait-online.service.d
  - systemctl daemon-reload
  - netplan apply
  - systemctl start ssh
  - systemctl start avahi-daemon
`

// portfwdTemplate is the cloud-init template for port forwarding
var portfwdTemplate = `#cloud-config
datasource_list: [ NoCloud, None ]
final_message: "CLOUD-INIT-READY"
hostname: {{.Hostname}}
fqdn: {{.Hostname}}.local
manage_etc_hosts: true
package_update: false
packages:
  - openssh-server
  - avahi-daemon
output:
  all: '| tee -a /var/log/cloud-init-output.log'
users:
  - name: {{.Username}}
    gecos: {{.Username}}
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: [sudo]
    shell: /bin/bash
    lock_passwd: false
    plain_text_passwd: '{{.Password}}'
ssh_pwauth: true
chpasswd:
  list: |
    {{.Username}}:{{.Password}}
  expire: false
write_files:
  - path: /etc/ssh/sshd_config.d/50-cloud-init.conf
    owner: root:root
    permissions: '0644'
    content: |
      PasswordAuthentication yes
      PubkeyAuthentication yes
      PermitRootLogin prohibit-password
      UseDNS no
      GSSAPIAuthentication no
runcmd:
  - |
    hostnamectl set-hostname {{.Hostname}} || echo "{{.Hostname}}" > /etc/hostname
    hostname {{.Hostname}}
  - echo "{{.Username}}:{{.Password}}" | chpasswd
  - passwd -u {{.Username}}
  - systemctl enable ssh
  - systemctl enable avahi-daemon
  - |
    if [ ! -f /etc/ssh/.keys_generated ]; then
      ssh-keygen -A
      touch /etc/ssh/.keys_generated
    fi
  - systemctl start ssh
  - systemctl start avahi-daemon
`

// TemplateData contains the data to populate cloud-init templates
type TemplateData struct {
	Hostname string
	Username string
	Password string
	StaticIP string
}

// Generator handles cloud-init configuration generation
type Generator struct{}

// NewGenerator creates a new cloud-init generator
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateSeedISO creates a cloud-init seed ISO for the VM
func (g *Generator) GenerateSeedISO(vm *models.VM, vmDir string) (string, error) {
	// Create template data
	data := TemplateData{
		Hostname: vm.Name,
		Username: vm.Username,
		Password: vm.Password,
		StaticIP: vm.StaticIP,
	}

	// Select appropriate template based on network type
	var tmplContent string
	if vm.NetworkType == models.NetworkBridge {
		tmplContent = bridgeTemplate
	} else {
		tmplContent = portfwdTemplate
	}

	// Parse and execute template
	tmpl, err := template.New("cloud-init").Parse(tmplContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse cloud-init template: %w", err)
	}

	var userDataBuf bytes.Buffer
	if err := tmpl.Execute(&userDataBuf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Create metadata
	metaData := map[string]string{
		"instance-id":    vm.Name,
		"local-hostname": vm.Name,
	}

	metaDataJSON, err := json.Marshal(metaData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Create temporary directory for cloud-init files
	tmpDir, err := os.MkdirTemp("", "cloud-init-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write user-data and meta-data files
	userDataPath := filepath.Join(tmpDir, "user-data")
	if err := os.WriteFile(userDataPath, userDataBuf.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("failed to write user-data: %w", err)
	}

	metaDataPath := filepath.Join(tmpDir, "meta-data")
	if err := os.WriteFile(metaDataPath, metaDataJSON, 0644); err != nil {
		return "", fmt.Errorf("failed to write meta-data: %w", err)
	}

	// Save cloud-init config to VM directory for debugging
	debugUserDataPath := filepath.Join(vmDir, "cloud-init-user-data.yaml")
	if err := os.WriteFile(debugUserDataPath, userDataBuf.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("failed to save debug user-data: %w", err)
	}

	debugMetaDataPath := filepath.Join(vmDir, "cloud-init-meta-data.json")
	if err := os.WriteFile(debugMetaDataPath, metaDataJSON, 0644); err != nil {
		return "", fmt.Errorf("failed to save debug meta-data: %w", err)
	}

	// Create ISO using hdiutil (macOS)
	seedISO := filepath.Join(vmDir, "seed.iso")

	cmd := exec.Command("hdiutil", "makehybrid",
		"-o", seedISO,
		"-hfs",
		"-joliet",
		"-iso",
		"-default-volume-name", "cidata",
		tmpDir,
	)

	if err := cmd.Run(); err != nil {
		// Fallback to mkisofs if available
		return g.createISOWithMkisofs(tmpDir, seedISO)
	}

	return seedISO, nil
}

// createISOWithMkisofs creates ISO using mkisofs or genisoimage as fallback
func (g *Generator) createISOWithMkisofs(srcDir, destISO string) (string, error) {
	// Try mkisofs first
	cmd := exec.Command("mkisofs", "-V", "cidata", "-o", destISO, "-J", "-r", srcDir)
	if err := cmd.Run(); err == nil {
		return destISO, nil
	}

	// Try genisoimage
	cmd = exec.Command("genisoimage", "-V", "cidata", "-o", destISO, "-J", "-r", srcDir)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create ISO with mkisofs/genisoimage: %w", err)
	}

	return destISO, nil
}
