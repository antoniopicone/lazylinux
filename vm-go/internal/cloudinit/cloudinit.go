package cloudinit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

type Config struct {
	Hostname string
	Username string
	Password string
	NetType  string
	StaticIP string
}

const userDataTemplate = `#cloud-config
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
{{if eq .NetType "bridge"}}
  - path: /etc/netplan/01-netcfg.yaml
    content: |
      network:
        version: 2
        renderer: networkd
        ethernets:
          enp0s1:
{{if .StaticIP}}
            addresses:
              - {{.StaticIP}}/24
            routes:
              - to: default
                via: 192.168.105.1
            nameservers:
              addresses: [8.8.8.8, 1.1.1.1]
            dhcp4: no
            dhcp6: no
{{else}}
            dhcp4: yes
            dhcp6: no
            nameservers:
              addresses: [8.8.8.8, 1.1.1.1]
{{end}}
{{end}}
runcmd:
  - hostnamectl set-hostname {{.Hostname}} || echo "{{.Hostname}}" > /etc/hostname
  - hostname {{.Hostname}}
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
{{if eq .NetType "bridge"}}
  - netplan apply
{{end}}
`

func CreateISO(vmDir string, cfg Config) (string, error) {
	tmpDir, err := os.MkdirTemp("", "cloudinit")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	// Generate user-data
	tmpl, err := template.New("user-data").Parse(userDataTemplate)
	if err != nil {
		return "", err
	}

	userDataFile, err := os.Create(filepath.Join(tmpDir, "user-data"))
	if err != nil {
		return "", err
	}

	if err := tmpl.Execute(userDataFile, cfg); err != nil {
		userDataFile.Close()
		return "", err
	}
	userDataFile.Close()

	// Generate meta-data
	metaData := fmt.Sprintf(`{"instance-id": "%s", "local-hostname": "%s"}`, cfg.Hostname, cfg.Hostname)
	if err := os.WriteFile(filepath.Join(tmpDir, "meta-data"), []byte(metaData), 0644); err != nil {
		return "", err
	}

	// Save for debugging
	os.WriteFile(filepath.Join(vmDir, "cloud-init-user-data.yaml"), []byte(userDataTemplate), 0644) // Saving template for now, ideally save rendered

	seedISO := filepath.Join(vmDir, "seed.iso")

	// Try hdiutil first (macOS)
	cmd := exec.Command("hdiutil", "makehybrid", "-o", seedISO, "-hfs", "-joliet", "-iso", "-default-volume-name", "cidata", tmpDir)
	if err := cmd.Run(); err == nil {
		return seedISO, nil
	}

	// Fallback to mkisofs
	cmd = exec.Command("mkisofs", "-V", "cidata", "-o", seedISO, "-J", "-r", tmpDir)
	if err := cmd.Run(); err == nil {
		return seedISO, nil
	}

	return "", fmt.Errorf("failed to create cloud-init ISO: install mkisofs or use macOS")
}
