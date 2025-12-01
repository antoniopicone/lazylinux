# ants.os - Post-Installation Script

**ants.os** è una distribuzione Linux minimale basata su Debian 13, installabile tramite uno script su qualsiasi VM Debian esistente.

## Quick Start

### 1. Crea una VM Debian minimale

```bash
# Usa lo script vm per creare una VM Debian
vm create --name ants-vm --user admin --pass test

# Aspetta che la VM sia pronta
vm wait ants-vm

# Connettiti via SSH
ssh admin@ants-vm.local
```

### 2. Installa ants.os

```bash
# Dentro la VM, scarica e esegui lo script
curl -fsSL https://raw.githubusercontent.com/antoniopicone/ants-os/main/install.sh | sudo bash

# Oppure, se hai lo script localmente:
# Copia lo script nella VM
scp install.sh admin@ants-vm.local:~

# SSH nella VM ed esegui
ssh admin@ants-vm.local
sudo ./install.sh
```

### 3. Riavvia

```bash
sudo reboot
```

## Cosa Viene Installato

### Core Tools
- **vim** - Editor di testo
- **git** - Version control
- **curl, wget** - Download utilities
- **htop** - Process monitor
- **build-essential** - Compilatori e tools

### System
- **sudo** - Privilege escalation
- **openssh-server** - SSH server
- **avahi-daemon** - mDNS (`.local` domains)

### Networking
- **NetworkManager** - Network configuration
- **Bluetooth** - Bluetooth support

### Display
- **GDM3** - Display manager (minimal, no full GNOME)
- **Xorg** - X Window System
- **Kitty** - GPU-accelerated terminal emulator

## Configurazioni Applicate

### NetworkManager
- Gestisce tutte le interfacce di rete
- systemd-networkd disabilitato
- DNS configurato

### GDM3
- Wayland abilitato
- Configurato per 120Hz (se supportato dall'hardware)
- Nessun auto-login

### Kitty Terminal
- Color scheme: Monokai
- Performance ottimizzata
- Configurazione in `~/.config/kitty/kitty.conf`

### User Configuration
- Utente corrente aggiunto a gruppo `sudo`
- NOPASSWD sudo abilitato
- Configurazioni copiate in home directory

### Branding
- Hostname: `ants-os`
- Custom `/etc/issue` e `/etc/motd`
- Welcome message personalizzato

## Utilizzo

### Dopo l'installazione

```bash
# Riavvia
sudo reboot

# Login con le tue credenziali
# Il sistema ora è ants.os!

# Verifica versione
cat /etc/motd

# Gestione rete
nmcli device status
nmcli device wifi list
nmcli device wifi connect "SSID" password "PASSWORD"

# Installare software aggiuntivo
sudo apt update
sudo apt install firefox-esr  # Browser
sudo apt install python3      # Python
```

### Configurare 120Hz

Se il tuo hardware supporta 120Hz:

```bash
# Via GUI (se hai desktop environment)
# Settings → Displays → 120Hz

# Via xrandr
xrandr --output Virtual-1 --mode 1920x1080 --rate 120
```

## Script Locale

Se vuoi usare lo script localmente invece di scaricarlo:

```bash
# 1. Crea VM
vm create --name ants-vm --user admin --pass test

# 2. Copia script
scp ants-os/install.sh admin@ants-vm.local:~

# 3. SSH e installa
ssh admin@ants-vm.local
sudo ./install.sh

# 4. Riavvia
sudo reboot
```

## Personalizzazione

### Modificare lo Script

Puoi modificare `install.sh` per:

- Aggiungere/rimuovere pacchetti
- Cambiare configurazioni
- Aggiungere script personalizzati

Esempio - aggiungere pacchetti:

```bash
# Trova la sezione "Install core tools"
apt-get install -y -qq \
    vim \
    git \
    your-package-here \  # Aggiungi qui
    curl \
    wget
```

### Aggiungere Desktop Environment

Lo script installa solo GDM3 senza desktop completo. Per aggiungere:

```bash
# GNOME minimal
sudo apt install gnome-core

# XFCE (lightweight)
sudo apt install xfce4

# KDE Plasma
sudo apt install kde-plasma-desktop
```

## Troubleshooting

### NetworkManager non funziona

```bash
# Restart service
sudo systemctl restart NetworkManager

# Check status
sudo systemctl status NetworkManager

# Check devices
nmcli device status
```

### GDM3 non parte

```bash
# Check logs
sudo journalctl -u gdm3

# Restart
sudo systemctl restart gdm3

# Fallback to console
sudo systemctl set-default multi-user.target
```

### Kitty non funziona

```bash
# Fallback a terminal alternativo
sudo apt install terminator

# O usa il terminal di sistema
xterm
```

## Architettura

```
Debian 13 (Trixie) Base
    ↓
install.sh
    ↓
ants.os
```

Lo script:
1. Aggiorna il sistema
2. Installa pacchetti richiesti
3. Configura servizi
4. Applica branding
5. Abilita servizi all'avvio

## Vantaggi vs ISO Custom

| Aspetto | ISO Custom | Script Post-Install |
|---------|------------|---------------------|
| Setup | Complesso | Semplice |
| Manutenzione | Difficile | Facile |
| Testing | Lento | Veloce |
| Personalizzazione | Limitata | Illimitata |
| Aggiornamenti | Rebuild ISO | Modifica script |

## Esempi d'Uso

### Installazione Rapida

```bash
# One-liner completo
vm create --name ants && \
vm wait ants && \
ssh admin@ants.local "curl -fsSL https://raw.../install.sh | sudo bash" && \
vm stop ants && \
vm start ants
```

### Installazione con Personalizzazioni

```bash
# Modifica lo script
vim ants-os/install.sh

# Aggiungi i tuoi pacchetti
# Modifica configurazioni

# Copia e installa
scp ants-os/install.sh admin@ants-vm.local:~
ssh admin@ants-vm.local "sudo ./install.sh"
```

### Creare Template

```bash
# 1. Installa ants.os su VM
# 2. Configura come preferisci
# 3. Spegni VM
# 4. Clona il disco

# Ora hai un template ants.os riutilizzabile!
```

## Requisiti Sistema

### Minimi
- Debian 13 (Trixie) o Debian 12 (Bookworm)
- 512 MB RAM
- 5 GB disk
- Connessione Internet

### Raccomandati
- 2 GB RAM
- 20 GB disk
- CPU dual-core

## License

MIT License - Usa, modifica, distribuisci liberamente

## Credits

- Basato su Debian GNU/Linux
- Ispirato da distribuzioni minimali come Arch e Alpine
- Creato per sviluppatori e power users

---

**Made with 🐜 by ants.os team**
