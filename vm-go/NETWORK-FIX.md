# Risoluzione Problema: VM Non Esce in Rete

## Problema

La VM ha un IP nella range 192.168.105.0/24 ed è accessibile via SSH dal Mac, ma non riesce ad accedere a Internet.

## Diagnosi

### Verifiche Effettuate

1. **IP Forwarding**: ✅ Abilitato
   ```bash
   sysctl net.inet.ip.forwarding
   # Output: net.inet.ip.forwarding: 1
   ```

2. **socket_vmnet**: ✅ In esecuzione
   ```bash
   ps aux | grep socket_vmnet
   # Output: /opt/homebrew/opt/socket_vmnet/bin/socket_vmnet --vmnet-gateway=192.168.105.1
   ```

3. **Bridge Interface**: ✅ Configurato
   ```bash
   ifconfig bridge100
   # Output: inet 192.168.105.1 netmask 0xffffff00
   ```

## Causa del Problema

**socket_vmnet di default NON abilita il NAT**. La rete 192.168.105.0/24 è isolata e non può raggiungere Internet senza NAT o routing esplicito.

## Soluzioni

### Soluzione 1: Usare vmnet-shared (Raccomandato)

socket_vmnet supporta due modalità:
- `vmnet-bridged`: Bridge diretto (nessun NAT)
- `vmnet-shared`: NAT abilitato (come VMware Fusion)

#### Modifica Configurazione socket_vmnet

```bash
# 1. Ferma socket_vmnet
sudo brew services stop socket_vmnet

# 2. Modifica il comando di avvio per usare vmnet-shared
# Edita il file di servizio o usa questo comando:
sudo /opt/homebrew/opt/socket_vmnet/bin/socket_vmnet \
  --vmnet-mode=shared \
  --vmnet-gateway=192.168.105.1 \
  /opt/homebrew/var/run/socket_vmnet &

# 3. Verifica che sia attivo
ps aux | grep socket_vmnet
```

#### Aggiorna vm-go per Usare vmnet-shared

Modifica `cmd/create.go` per passare l'opzione corretta a socket_vmnet_client:

```go
// In futuro, aggiungere supporto per --vmnet-mode=shared
```

### Soluzione 2: Configurare NAT Manualmente con pfctl

Se vuoi mantenere vmnet-bridged, devi configurare NAT manualmente:

```bash
# 1. Crea file di configurazione pfctl
sudo tee /etc/pf.anchors/vm.nat << 'EOF'
# NAT per VMs su bridge100
nat on en0 from 192.168.105.0/24 to any -> (en0)
EOF

# 2. Aggiungi anchor al file pf.conf principale
sudo tee -a /etc/pf.conf << 'EOF'
nat-anchor "vm.nat"
load anchor "vm.nat" from "/etc/pf.anchors/vm.nat"
EOF

# 3. Ricarica pfctl
sudo pfctl -f /etc/pf.conf
sudo pfctl -e

# 4. Verifica NAT
sudo pfctl -s nat
```

**Nota**: Sostituisci `en0` con la tua interfaccia di rete principale (usa `ifconfig` per trovarla).

### Soluzione 3: Usare Port Forwarding (Workaround)

Se non hai bisogno di accesso diretto alla rete, usa port forwarding:

```bash
./vm-go create --name myvm --net-type portfwd
```

Questo usa QEMU user networking che ha NAT integrato.

## Verifica della Soluzione

Dopo aver applicato una delle soluzioni, testa dalla VM:

```bash
# SSH nella VM
ssh anto@192.168.105.193

# Test connettività Internet
ping -c 3 8.8.8.8          # Test IP
ping -c 3 google.com       # Test DNS
curl -I https://google.com # Test HTTP
```

## Implementazione Raccomandata per vm-go

### Opzione 1: Aggiungere Flag --vmnet-mode

```go
// cmd/create.go
var vmnetMode string

createCmd.Flags().StringVar(&vmnetMode, "vmnet-mode", "shared", "VMNet mode (shared, bridged)")

// Quando avvii socket_vmnet_client, passa il parametro
```

### Opzione 2: Documentare Setup Iniziale

Aggiungere al README:

```markdown
## Setup Bridge Networking con NAT

Per abilitare l'accesso a Internet dalle VMs:

1. Ferma socket_vmnet:
   ```bash
   sudo brew services stop socket_vmnet
   ```

2. Avvia con modalità shared:
   ```bash
   sudo /opt/homebrew/opt/socket_vmnet/bin/socket_vmnet \
     --vmnet-mode=shared \
     --vmnet-gateway=192.168.105.1 \
     /opt/homebrew/var/run/socket_vmnet &
   ```

3. Rendi permanente modificando il servizio brew
```

## Confronto Modalità

| Modalità | NAT | Accesso Internet | Accesso LAN | Uso |
|----------|-----|------------------|-------------|-----|
| `shared` | ✅ | ✅ | ❌ | Sviluppo, test |
| `bridged` | ❌ | ⚠️ Richiede config | ✅ | Produzione, server |
| `portfwd` | ✅ | ✅ | ❌ | Semplice, isolato |

## Prossimi Passi

1. ⏳ Aggiungere comando `setup-bridge` che configura automaticamente socket_vmnet con NAT
2. ⏳ Aggiungere flag `--vmnet-mode` al comando `create`
3. ⏳ Documentare nel README le diverse modalità di networking
4. ⏳ Creare script di diagnostica per verificare configurazione NAT

## Soluzione Rapida (Temporanea)

Per ora, la soluzione più rapida è:

```bash
# Ferma il servizio attuale
sudo brew services stop socket_vmnet

# Avvia manualmente con shared mode
sudo /opt/homebrew/opt/socket_vmnet/bin/socket_vmnet \
  --vmnet-mode=shared \
  --vmnet-gateway=192.168.105.1 \
  /opt/homebrew/var/run/socket_vmnet &

# Ricrea la VM
./vm-go delete anto3 --force
./vm-go create --name anto3 --user anto --pass test

# Testa
ssh anto@<IP>
ping 8.8.8.8
```
