#!/bin/bash

# ants.os Installation Script
# Transforms a minimal Debian VM into ants.os
# Usage: curl -fsSL https://raw.githubusercontent.com/.../install.sh | sudo bash
#    or: sudo ./install.sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m'

VERSION="1.0"

log() {
    echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $*"
}

success() {
    echo -e "${GREEN}✔${NC} $*"
}

error() {
    echo -e "${RED}✗${NC} $*" >&2
    exit 1
}

warning() {
    echo -e "${YELLOW}⚠${NC} $*"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    error "Please run as root (use sudo)"
fi

# Banner
echo ""
echo "╔═══════════════════════════════════════╗"
echo "║       ants.os Installation v${VERSION}      ║"
echo "║   Minimal Debian-based Distribution   ║"
echo "╚═══════════════════════════════════════╝"
echo ""

# Detect architecture
ARCH=$(uname -m)
if [ "$ARCH" = "aarch64" ]; then
    ARCH="arm64"
elif [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
fi

log "Detected architecture: $ARCH"

# Update system
log "Updating package lists..."
apt-get update -qq

log "Upgrading existing packages..."
apt-get upgrade -y -qq

# Install core tools
log "Installing core development tools..."
apt-get install -y -qq \
    vim \
    git \
    curl \
    wget \
    htop \
    build-essential \
    software-properties-common

success "Core tools installed"

# Install system utilities
log "Installing system utilities..."
apt-get install -y -qq \
    sudo \
    openssh-server \
    avahi-daemon \
    locales-all

success "System utilities installed"

# Install networking
log "Installing NetworkManager..."
apt-get install -y -qq \
    network-manager \
    network-manager-gnome

# Disable systemd-networkd in favor of NetworkManager
systemctl disable systemd-networkd 2>/dev/null || true
systemctl mask systemd-networkd 2>/dev/null || true
systemctl enable NetworkManager
systemctl start NetworkManager

success "NetworkManager configured"

# Install Bluetooth
log "Installing Bluetooth support..."
apt-get install -y -qq \
    bluetooth \
    bluez \
    bluez-tools

systemctl enable bluetooth
systemctl start bluetooth

success "Bluetooth installed"

# Install minimal display manager and X
log "Installing display manager (GDM3)..."
apt-get install -y -qq \
    gdm3 \
    xorg \
    xinit \
    x11-xserver-utils

# Don't install full GNOME, just GDM3
systemctl enable gdm3

success "Display manager installed"

# Install Kitty terminal
log "Installing Kitty terminal..."
apt-get install -y -qq kitty

# Configure Kitty
mkdir -p /etc/skel/.config/kitty
cat > /etc/skel/.config/kitty/kitty.conf << 'EOF'
# ants.os Kitty Configuration

# Font
font_family      monospace
font_size        11.0

# Performance
repaint_delay    10
input_delay      3
sync_to_monitor  yes

# Window
remember_window_size  yes
initial_window_width  1024
initial_window_height 768

# Tab bar
tab_bar_edge top
tab_bar_style powerline

# Color scheme (Monokai)
background #272822
foreground #f8f8f2
cursor #f8f8f2

# Black
color0 #272822
color8 #75715e

# Red
color1 #f92672
color9 #f92672

# Green
color2  #a6e22e
color10 #a6e22e

# Yellow
color3  #f4bf75
color11 #f4bf75

# Blue
color4  #66d9ef
color12 #66d9ef

# Magenta
color5  #ae81ff
color13 #ae81ff

# Cyan
color6  #a1efe4
color14 #a1efe4

# White
color7  #f8f8f2
color15 #f9f8f5
EOF

success "Kitty terminal configured"

# Configure NetworkManager
log "Configuring NetworkManager..."
cat > /etc/NetworkManager/NetworkManager.conf << 'EOF'
[main]
plugins=ifupdown,keyfile
dns=default

[ifupdown]
managed=true

[device]
wifi.scan-rand-mac-address=no
EOF

success "NetworkManager configured"

# Configure GDM3 for 120Hz
log "Configuring GDM3..."
mkdir -p /etc/gdm3

cat > /etc/gdm3/custom.conf << 'EOF'
[daemon]
WaylandEnable=true
AutomaticLoginEnable=false

[security]

[xdmcp]

[chooser]

[debug]
Enable=false
EOF

# Create monitors configuration for 120Hz
mkdir -p /var/lib/gdm3/.config
cat > /var/lib/gdm3/.config/monitors.xml << 'EOF'
<monitors version="2">
  <configuration>
    <logicalmonitor>
      <x>0</x>
      <y>0</y>
      <scale>1</scale>
      <primary>yes</primary>
      <monitor>
        <monitorspec>
          <connector>Virtual-1</connector>
          <vendor>unknown</vendor>
          <product>unknown</product>
          <serial>unknown</serial>
        </monitorspec>
        <mode>
          <width>1920</width>
          <height>1080</height>
          <rate>120</rate>
        </mode>
      </monitor>
    </logicalmonitor>
  </configuration>
</monitors>
EOF

chown -R Debian-gdm:Debian-gdm /var/lib/gdm3/.config 2>/dev/null || true

success "GDM3 configured for 120Hz"

# Configure user (if not root)
CURRENT_USER="${SUDO_USER:-$USER}"
if [ "$CURRENT_USER" != "root" ] && [ -n "$CURRENT_USER" ]; then
    log "Configuring user: $CURRENT_USER"
    
    # Add to sudo group
    usermod -aG sudo "$CURRENT_USER"
    
    # Configure NOPASSWD sudo
    echo "$CURRENT_USER ALL=(ALL) NOPASSWD:ALL" > "/etc/sudoers.d/$CURRENT_USER"
    chmod 0440 "/etc/sudoers.d/$CURRENT_USER"
    
    # Copy Kitty config to user home
    if [ -d "/home/$CURRENT_USER" ]; then
        mkdir -p "/home/$CURRENT_USER/.config/kitty"
        cp /etc/skel/.config/kitty/kitty.conf "/home/$CURRENT_USER/.config/kitty/"
        chown -R "$CURRENT_USER:$CURRENT_USER" "/home/$CURRENT_USER/.config"
    fi
    
    success "User $CURRENT_USER configured"
fi

# Create ants.os branding
log "Creating ants.os branding..."

cat > /etc/issue << 'EOF'
ants.os \n \l

Welcome to ants.os - A minimal Debian-based distribution

EOF

cat > /etc/motd << 'EOF'
╔═══════════════════════════════════════╗
║          Welcome to ants.os           ║
║   A minimal Debian-based distribution ║
╚═══════════════════════════════════════╝

System Information:
  - Terminal: kitty
  - Network: NetworkManager
  - Display: GDM3 (120Hz capable)

Quick Start:
  - Update system: sudo apt update && sudo apt upgrade
  - Install packages: sudo apt install <package>
  - Network: nmcli device wifi list

Documentation: https://github.com/antoniopicone/ants-os

EOF

# Set hostname
log "Setting hostname to ants-os..."
hostnamectl set-hostname ants-os || echo "ants-os" > /etc/hostname

cat > /etc/hosts << 'EOF'
127.0.0.1       localhost
127.0.1.1       ants-os.local ants-os

::1             localhost ip6-localhost ip6-loopback
ff02::1         ip6-allnodes
ff02::2         ip6-allrouters
EOF

success "Branding configured"

# Enable services
log "Enabling services..."
systemctl enable ssh
systemctl enable avahi-daemon
systemctl enable NetworkManager
systemctl enable bluetooth
systemctl enable gdm3

success "Services enabled"

# Clean up
log "Cleaning up..."
apt-get autoremove -y -qq
apt-get clean

success "Cleanup complete"

# Summary
echo ""
echo "╔═══════════════════════════════════════╗"
echo "║   ants.os Installation Complete! ✔    ║"
echo "╚═══════════════════════════════════════╝"
echo ""
echo "Installed packages:"
echo "  ✔ vim, git, curl, wget, htop"
echo "  ✔ openssh-server, avahi-daemon"
echo "  ✔ NetworkManager, Bluetooth"
echo "  ✔ GDM3, Xorg, Kitty terminal"
echo ""
echo "Next steps:"
echo "  1. Reboot the system: sudo reboot"
echo "  2. Login with your user credentials"
echo "  3. Enjoy ants.os!"
echo ""
echo "For 120Hz display:"
echo "  Settings → Displays → Select 120Hz refresh rate"
echo ""
