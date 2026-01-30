# Deploy via LXC

This guide provides a streamlined workflow for manually deploying Redroid using native LXC. For automated deployment, use the [redway](https://github.com/jimed-rand/redway) tool.

## Prerequisites

Ensure your distribution has the required packages installed.

| Distribution | Command |
| --- | --- |
| **Ubuntu/Debian** | `sudo apt install lxc-utils skopeo umoci jq linux-modules-extra-$(uname -r)` |
| **Arch Linux** | `sudo pacman -S lxc skopeo umoci jq linux-headers` |
| **Fedora** | `sudo dnf install lxc lxc-templates skopeo umoci jq kernel-modules-extra` |
| **openSUSE** | `sudo zypper install lxc skopeo umoci jq` |
| **Gentoo** | `emerge -av app-containers/lxc app-containers/skopeo app-containers/umoci app-misc/jq` |

### Step 1: Environment Preparation

> [!IMPORTANT]
> Redroid requires **binder** support. Check if it's available or load the module:

```bash
# Check for binder support
ls -la /dev/binder* /dev/binderfs 2>/dev/null || sudo modprobe binder_linux devices="binder,hwbinder,vndbinder"

# Verify LXC networking is active
sudo systemctl enable --now lxc-net
ip link show lxcbr0
```

> [!TIP]
> If you're on a newer kernel using `binderfs`, ensure it's mounted or that the devices are present in `/dev/`.

### Step 2: Create Container

First, patch the OCI template for better compatibility, then create the container:

```bash
# Fix OCI template compatibility
sudo sed -i 's/set -eu/set -u/g' /usr/share/lxc/templates/lxc-oci

# Create container (Android 14 example)
# Versions available: 11.0.0 through 16.0.0
sudo lxc-create -n redroid -t oci -- -u docker://redroid/redroid:14.0.0_64only-latest
```

### Step 3: Configure & Fixes

Merge persistence, Redroid settings, and networking workarounds into the container configuration:

```bash
# Create data directory
mkdir -p $HOME/data-redroid

# Clean and configure
sudo sed -i '/lxc.include/d' /var/lib/lxc/redroid/config
sudo tee -a /var/lib/lxc/redroid/config <<EOF
### Redroid Configuration
lxc.init.cmd = /init androidboot.hardware=redroid androidboot.redroid_gpu_mode=guest
lxc.apparmor.profile = unconfined
lxc.autodev = 1
lxc.autodev.tmpfs.size = 25000000
lxc.mount.entry = $HOME/data-redroid data none bind 0 0
EOF

# Networking workaround: Remove ipconfigstore
sudo rm -f /var/lib/lxc/redroid/rootfs/vendor/bin/ipconfigstore
```

### Step 4: Launch & Connect

Start the container and connect via ADB:

```bash
# Start container
sudo lxc-start -n redroid -l debug -o redroid.log

# Get IP and Wait for boot (approx 30-60s)
REDROID_IP=$(sudo lxc-info redroid -i | awk '{print $2}')
adb connect $REDROID_IP:5555
```

---

## Container Management

| Action | Command |
|--------|---------|
| **Status** | `sudo lxc-info redroid` |
| **Stop** | `sudo lxc-stop -k -n redroid` |
| **Restart** | `sudo lxc-stop -k -n redroid && sudo lxc-start -n redroid` |
| **Shell** | `sudo nsenter -t $(sudo lxc-info redroid -p \| awk '{print $2}') -a sh` |
| **Logs** | `tail -f redroid.log` |
| **Delete** | `sudo lxc-destroy -n redroid` |

## GPU Acceleration

To enable hardware acceleration, change `redroid_gpu_mode` to `host`:

```bash
sudo sed -i 's/redroid_gpu_mode=guest/redroid_gpu_mode=host/' /var/lib/lxc/redroid/config
```

> [!NOTE]
> `host` mode requires compatible GPU drivers (Mesa/Intel/AMD recommended).

## Troubleshooting

> [!CAUTION]
> **Container fails to start?**
> Check `redroid.log` and verify binder devices: `ls -la /dev/binder*`.

> [!WARNING]
> **No Network?**
> Ensure `lxcbr0` exists: `ip link show lxcbr0`. Restart if needed: `sudo systemctl restart lxc-net`.

> [!TIP]
> **ADB Connection Refused?**
> Android takes time to boot. Wait at least 60 seconds before connecting.

