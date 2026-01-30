# Deploy via LXC

This guide shows how to manually deploy a Redroid container using LXC.
For automated deployment, use the `redway` tool instead.

## Prerequisites by Distribution

### Ubuntu/Debian

```bash
# Install kernel module support (if needed)
sudo apt install linux-modules-extra-$(uname -r)

# Install LXC and OCI tools
sudo apt install lxc-utils skopeo umoci jq
```

### Arch Linux

```bash
# Install all required packages
sudo pacman -S linux-headers lxc skopeo umoci jq
```

### Fedora

```bash
# Install kernel modules and LXC tools
sudo dnf install kernel-modules-extra lxc lxc-templates skopeo umoci jq
```

### openSUSE

```bash
sudo zypper install lxc skopeo umoci jq
```

### Gentoo

```bash
emerge -av app-containers/lxc app-containers/skopeo app-containers/umoci app-misc/jq
```

## Step 1: Verify Binder Support

Redroid requires binder support. Different distributions handle this differently:

```bash
# Check for binder support (any of these indicates support)
ls -la /dev/binder* /dev/binderfs 2>/dev/null
lsmod | grep binder
cat /proc/filesystems | grep binder
```

### If binder is not found

```bash
# Try loading modules (common on Ubuntu/Debian/Fedora)
sudo modprobe binder_linux devices="binder,hwbinder,vndbinder"

# Verify
lsmod | grep binder
```

> **Note:** Some distributions (like newer Arch, some custom kernels) have binder support built-in or use binderfs. The container will work as long as /dev/binder* devices are available.

## Step 2: Verify LXC Networking

```bash
# Check if lxc-net service is running (systemd)
sudo systemctl status lxc-net

# If not running
sudo systemctl start lxc-net
sudo systemctl enable lxc-net

# Verify bridge exists
ip link show lxcbr0
```

## Step 3: Adjust OCI Template

```bash
# Fix OCI template for compatibility
sudo sed -i 's/set -eu/set -u/g' /usr/share/lxc/templates/lxc-oci
```

## Step 4: Create Redroid Container

```bash
# Create container from OCI image
# Available Android versions: 11.0.0, 12.0.0, 13.0.0, 14.0.0, 15.0.0, 16.0.0
sudo lxc-create -n redroid -t oci -- -u docker://redroid/redroid:14.0.0_64only-latest
```

## Step 5: Configure Container

```bash
# Create data directory for persistence
mkdir -p $HOME/data-redroid

# Remove problematic lxc.include line
sudo sed -i '/lxc.include/d' /var/lib/lxc/redroid/config

# Add Redroid-specific configuration
sudo tee -a /var/lib/lxc/redroid/config <<EOF
### Redroid Configuration
lxc.init.cmd = /init androidboot.hardware=redroid androidboot.redroid_gpu_mode=guest
lxc.apparmor.profile = unconfined
lxc.autodev = 1
lxc.autodev.tmpfs.size = 25000000
lxc.mount.entry = $HOME/data-redroid data none bind 0 0
EOF
```

## Step 6: Apply Networking Workaround

```bash
# Remove ipconfigstore to fix networking issues
sudo rm -f /var/lib/lxc/redroid/rootfs/vendor/bin/ipconfigstore
```

## Step 7: Start Container

```bash
# Start with debug logging
sudo lxc-start -l debug -o redroid.log -n redroid

# Check container status
sudo lxc-info redroid
```

## Step 8: Connect via ADB

```bash
# Get container IP
REDROID_IP=$(sudo lxc-info redroid -i | awk '{print $2}')

# Connect ADB
adb connect $REDROID_IP:5555

# Verify connection
adb devices
```

## Alternative: Direct Shell Access

```bash
# Get container PID and enter shell
sudo nsenter -t $(sudo lxc-info redroid -p | awk '{print $2}') -a sh
```

## Container Management

```bash
# Stop container
sudo lxc-stop -k -n redroid

# Restart container
sudo lxc-stop -k -n redroid && sudo lxc-start -n redroid

# Remove container
sudo lxc-destroy -n redroid

# List all containers
sudo lxc-ls -f
```

## GPU Mode Options

| Mode | Description |
|------|-------------|
| `guest` | Software rendering (default, most compatible) |
| `host` | GPU passthrough (requires compatible hardware) |

```bash
# Change to host GPU mode
sudo sed -i 's/redroid_gpu_mode=guest/redroid_gpu_mode=host/' /var/lib/lxc/redroid/config
```

## Troubleshooting

### Container won't start

```bash
# Check logs
cat redroid.log

# Verify binder support
ls -la /dev/binder* /dev/binderfs 2>/dev/null
```

### No binder device

```bash
# Try loading modules
sudo modprobe binder_linux devices="binder,hwbinder,vndbinder"
sudo modprobe ashmem_linux

# If modules don't exist, check if kernel has built-in support
cat /proc/filesystems | grep binder
```

### No network connectivity

```bash
# Restart LXC networking
sudo systemctl restart lxc-net

# Verify bridge
ip link show lxcbr0
```

### ADB connection refused

```bash
# Wait for Android to fully boot (30-60 seconds)
sleep 60
adb connect $REDROID_IP:5555
```
