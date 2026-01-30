# Redway - Native LXC Container Manager for Redroid

Redway is a lightweight Go-based **LXC container manager** for [redroid](https://github.com/remote-android/redroid-doc). It provides a streamlined CLI for managing Android containers using native LXC instead of Docker. Redway handles OCI image fetching via skopeo/umoci and automatically detects binder support across distributions. Designed for performance, it enables efficient Android virtualization on every Linux distribution with LXC support. Redway offers a robust, low-overhead environment for running Redroid with native speed and reliability.

## Key Features

- **Pure LXC Implementation** - Uses native LXC containers, not Docker
- **OCI Image Support** - Fetches images from OCI registries via skopeo/umoci
- **Cross-Distro Compatible** - Works on every Linux distribution with LXC support and binder support (binderfs/binder module)
- **Simple CLI Interface** - Easy-to-use command-line tool
- **ADB Integration** - Built-in ADB connection management
- **Persistent Storage** - Android data persists across container restarts
- **Lightweight Single Binary** - No runtime dependencies, just system tools

## What is Redroid?

[Redroid](https://github.com/remote-android/redroid-doc) (Remote Android) is a GPU-accelerated **AIC (Android In Container)** solution that allows you to run Android in containers with near-native performance. By sharing the host's kernel and leveraging hardware acceleration, Redroid provides a high-density, low-latency environment for Android virtualization without the overhead of a traditional hypervisor or emulator. It is commonly used for cloud gaming, automated testing, and virtual mobile infrastructure.

## Architecture

Redway uses a pure LXC-based approach:

```
┌─────────────────────────────────────────────────┐
│                   Redway CLI                    │
├─────────────────────────────────────────────────┤
│                                                 │
│  ┌─────────┐    ┌─────────┐    ┌─────────────┐ │
│  │ skopeo  │───►│  umoci  │───►│ lxc-create  │ │
│  │(fetch)  │    │(unpack) │    │  (OCI tpl)  │ │
│  └─────────┘    └─────────┘    └─────────────┘ │
│                                      │          │
│                              ┌───────▼────────┐ │
│                              │  LXC Container │ │
│                              │   (redroid)    │ │
│                              └────────────────┘ │
└─────────────────────────────────────────────────┘
```

**How it works:**

1. `skopeo` fetches OCI images from container registries
2. `umoci` unpacks the OCI image into a rootfs
3. `lxc-create` creates an LXC container using the OCI template
4. LXC manages the container lifecycle (start/stop/restart)

## Prerequisites

### System Requirements

- Linux kernel with binder support (binderfs or binder module via DKMS or KMP for openSUSE)
- LXC tools installed (lxc-utils)
- Root/sudo privileges
- OCI image (e.g. docker://redroid/redroid:16.0.0_64only-latest)
- ADB (for ADB connection management)
- Any Intel CPU with VT-x enabled or AMD CPU with SVM enabled
- Recommmended RAM are 8 GB or higher and Storage are 32 GB or higher

### Binder Support

Redroid requires **binder** kernel support both **binderfs** or **binder module** via DKMS or KMP for openSUSE. Redway will check for binder support during initialization and provide guidance if not detected.

#### Checking Binder Support

```bash
# Check for binder (any of these indicates support)
ls -la /dev/binder* /dev/binderfs 2>/dev/null
lsmod | grep binder
```

#### Loading Modules (if needed)

```bash
sudo modprobe binder_linux devices="binder,hwbinder,vndbinder"
lsmod | grep binder # Check if modules are loaded
```

#### Persistent Module Loading (if using modules)

# Create module configuration
sudo tee /etc/modules-load.d/binder.conf <<EOF
binder_linux
EOF

sudo tee /etc/modprobe.d/binder.conf <<EOF
options binder_linux devices="binder,hwbinder,vndbinder"
EOF

> **Note:** If your kernel has built-in binder support (common on Arch and
> custom kernels), you don't need to load modules. Redway will detect this
> automatically.

### Required Packages

Install these packages for your distribution:

#### Ubuntu/Debian

```bash
sudo apt update
sudo apt install -y lxc-utils skopeo umoci jq
```

#### Arch Linux

```bash
sudo pacman -S lxc skopeo umoci jq
```

#### Fedora

```bash
sudo dnf install -y lxc lxc-templates skopeo umoci jq
```

#### openSUSE

```bash
sudo zypper install lxc skopeo umoci jq
```

## Installation

### From Source

```bash
git clone https://github.com/jimed-rand/redway.git
cd redway
make build
sudo make install
```

### Manual Compilation

```bash
go build -ldflags="-s -w" -o redway .
sudo cp redway /usr/local/bin/
sudo chmod +x /usr/local/bin/redway
```

### Static Binary

```bash
make static
sudo cp redway /usr/local/bin/
```

## Quick Start

### 1. Load Kernel Modules

**Before using Redway, you must load the required kernel modules. This was already explained at [Binder Support](#binder-support) section.**

### 2. Initialize LXC Container

Using default image (Android 16):

```bash
sudo redway init
```

Using specific image (e.g. Android 16):

```bash
sudo redway init docker://redroid/redroid:16.0.0_64only-latest
```

Available redroid images:

- Android 16 (redroid/redroid:16.0.0-latest)
- Android 16 64bit only (redroid/redroid:16.0.0_64only-latest)
- Android 15 (redroid/redroid:15.0.0-latest)
- Android 15 64bit only (redroid/redroid:15.0.0_64only-latest)
- Android 14 (redroid/redroid:14.0.0-latest)
- Android 14 64bit only (redroid/redroid:14.0.0_64only-latest)
- Android 13 (redroid/redroid:13.0.0-latest)
- Android 13 64bit only (redroid/redroid:13.0.0_64only-latest)
- Android 12 (redroid/redroid:12.0.0-latest)
- Android 12 64bit only (redroid/redroid:12.0.0_64only-latest)
- Android 11 (redroid/redroid:11.0.0-latest)
- Android 10 (redroid/redroid:10.0.0-latest)
- Android 9 (redroid/redroid:9.0.0-latest)
- Android 8.1 (redroid/redroid:8.1.0-latest)

> **Note:** The `docker://` prefix is used by skopeo to identify OCI registry
> sources. This does **NOT** mean Docker is required - Redway uses native LXC
> containers.

### 3. Start Container

```bash
sudo redway start
```

### 4. Connect via ADB

Get connection information:

```bash
redway adb-connect
```

Connect:

```bash
adb connect <IP>:5555
```

### 5. Install Apps

```bash
adb install app.apk
```

## Usage

### Container Management

```bash
sudo redway init [image path]     # Initialize new LXC container
sudo redway start                 # Start container
sudo redway stop                  # Stop container
sudo redway restart               # Restart container
sudo redway remove                # Remove container
```

### Information & Debugging

```bash
redway status                # Show container status
redway list                  # List all LXC containers
redway log                   # Show container logs
redway adb-connect           # Display ADB connection info
```

### Direct Access

```bash
sudo redway shell            # Enter container via nsenter
```

## Configuration

Configuration is stored in `~/.config/redway/config.json`:

```json
{
  "container_name": "redroid",
  "image_url": "docker://redroid/redroid:16.0.0_64only-latest",
  "data_path": "/home/user/data-redroid",
  "log_file": "redroid.log",
  "gpu_mode": "guest",
  "initialized": true
}
```

### Configuration Options

| Option           | Description                   | Default                                         |
| ---------------- | ----------------------------- | ----------------------------------------------- |
| `container_name` | LXC container name            | `redroid`                                       |
| `image_url`      | OCI image reference           | `docker://redroid/redroid:16.0.0_64only-latest` |
| `data_path`      | Android data persistence path | `~/data-redroid`                                |
| `log_file`       | Log file location             | `redroid.log`                                   |
| `gpu_mode`       | GPU rendering mode            | `guest`                                         |

### GPU Modes

- `guest` - Software rendering (default, most compatible)
- `host` - GPU passthrough (requires compatible GPU)

## Project Structure

```
redway/
├── main.go                  # Entry point
├── go.mod                   # Go module definition
├── Makefile                 # Build system
├── cmd/                     # CLI commands
│   ├── command.go          # Command dispatcher
│   └── usage.go            # Help text
└── pkg/                     # Core packages
    ├── config/             # Configuration management
    │   └── config.go
    ├── container/          # LXC container operations
    │   ├── initializer.go  # Container setup
    │   └── manager.go      # Lifecycle management
    └── utils/              # Utility functions
        ├── adb.go          # ADB integration
        ├── shell.go        # Shell access
        ├── status.go       # Status reporting
        └── log.go          # Log viewing
```

## Directory Structure

```
~/.config/redway/
└── config.json              # Redway configuration

~/data-redroid/              # Android persistent data
├── data/                    # App data
└── system/                  # System data

/var/lib/lxc/redroid/        # LXC container
├── config                   # Container configuration
└── rootfs/                  # Android filesystem
```

## LXC Integration Details

Redway leverages LXC's OCI template to create containers from OCI images:

### Container Creation Flow

1. **Fetch Image** - `skopeo` downloads the OCI image from registry
2. **Unpack Image** - `umoci` extracts the image to a rootfs
3. **Create Container** - `lxc-create -t oci` creates the LXC container
4. **Configure** - Redway adjusts LXC config for Android compatibility
5. **Apply Workarounds** - Network and other Android-specific fixes

### LXC Configuration

Redway configures LXC containers with:

```
lxc.init.cmd = /init androidboot.hardware=redroid androidboot.redroid_gpu_mode=guest
lxc.apparmor.profile = unconfined
lxc.autodev = 1
lxc.autodev.tmpfs.size = 25000000
lxc.mount.entry = /home/user/data-redroid data none bind 0 0
```

## Troubleshooting

### Container won't start due to missing kernel modules or something else

Make sure that you have loaded the required kernel modules. This was already explained at [Binder Support](#binder-support) section. If you think that's something else, check at `redway log`.

### No IP address found

You need to check LXC networking by running:

```bash
sudo lxc-info redroid
```

If you don't see an IP address, try to restart the container:

```bash
sudo redway restart
```

### ADB connection fails

```bash
adb kill-server
adb start-server
adb connect <IP>:5555
```

### Permission issues

Always use sudo for container operations:

```bash
sudo redway start
```

## Advanced Usage

### Custom Data Directory

Edit `~/.config/redway/config.json`:

```json
{
  "data_path": "/mnt/storage/android-data"
}
```

### Multiple Containers

Edit container name in config before initializing:

```json
{
  "container_name": "redroid-test"
}
```

### Direct LXC Access

This step is only for advanced users only. You can access the container directly using `nsenter` or `lxc-attach`.

```bash
# Using nsenter
PID=$(sudo lxc-info redroid -p | awk '{print $2}')
sudo nsenter -t $PID -a sh

# Using lxc-attach
sudo lxc-attach -n redroid

# List LXC containers
sudo lxc-ls -f
```

## Distribution-Specific Notes

### Ubuntu/Debian

- **Kernel Modules**: Ensure `linux-modules-extra-$(uname -r)` is installed to provide necessary drivers for containerized environments.
- **Networking**: The `lxc-net` service is usually pre-configured. If `lxcbr0` is missing, run `sudo systemctl enable --now lxc-net`.
- **Dependencies**: `skopeo` and `umoci` are available in standard repositories since Ubuntu 20.04+.

### Arch Linux

- **Services**: You must manually enable the networking bridge: `sudo systemctl enable --now lxc-net`.
- **Kernel**: If using the LTS kernel, ensure `linux-lts-headers` are installed.
- **Dependencies**: Install `skopeo` and `umoci` from the official repositories.

### Fedora

- **SELinux**: Fedora's strict SELinux policy may prevent Android's `init` from mounting filesystems. You may need to set SELinux to permissive mode (`setenforce 0`) or use `lxc.apparmor.profile = unconfined`.
- **Cgroups**: Fedora uses cgroup v2 by default. If the container fails to start, you might need to revert to cgroup v1 using the kernel parameter `systemd.unified_cgroup_hierarchy=0`.
- **Kernel**: Install `kernel-modules-extra` to ensure binder support is available.

### openSUSE

- **Firewall**: `firewalld` often blocks the LXC bridge. Add the interface to the trusted zone: `sudo firewall-cmd --zone=trusted --add-interface=lxcbr0 --permanent`.
- **Subuid/Subgid**: Ensure your user has entries in `/etc/subuid` and `/etc/subgid` even when running with sudo to avoid mapping issues.

## Known Limitations

- **Kernel Requirements**: Requires a kernel with `binder` and `ashmem` support (or `binderfs`). Most modern desktop kernels include these, but stripped-down VPS kernels may not.
- **Root Access**: Due to the way LXC manages network bridges and device nodes (`/dev/binder`, `/dev/ashmem`), `sudo` is required for most lifecycle operations.
- **GPU Passthrough**: `gpu_mode: host` requires the host to have compatible Vulkan/EGL drivers. Intel and AMD (Mesa) generally work better than Nvidia in this specific LXC setup.
- **Architecture**: Ensure the image architecture (e.g., `64only`) matches your hardware capabilities.

## Why LXC Instead of Docker?

Redway uses LXC for several reasons:

1. **Lightweight** - No daemon required, direct kernel integration
2. **System Containers** - LXC is designed for full system containers like
   Android
3. **Fine-grained Control** - Direct access to container configuration
4. **No Extra Services** - No Docker daemon overhead
5. **Native Linux** - Uses standard Linux container primitives

## Contributing

Contributions to Redway are warmly welcomed and play a crucial role in the project's evolution. We invite developers of all skill levels to participate in enhancing our LXC-based Android management tool. To ensure the project remains robust and maintainable, we ask that all contributions adhere to a set of core guidelines. First, please ensure that all code follows standard Go idioms and formatting; consistency in the codebase allows for easier peer reviews and long-term stability. Second, maintaining cross-distribution compatibility is a top priority for Redway. Your changes should be verified across multiple Linux environments to ensure that users on Ubuntu, Arch, Fedora, and other distributions experience the same reliable performance. Third, documentation is essential for a project with this level of technical depth.

If your contribution introduces new configuration options, commands, or architectural changes, please update the README and relevant documentation files accordingly. Fourth, we emphasize clean system integration: please avoid hardcoded kernel module loading or distribution-specific assumptions, as Redway aims to be a flexible wrapper that respects the host's configuration. Beyond these technical requirements, we encourage open communication. If you find a bug or have a vision for a new feature, opening an issue is the best way to start a conversation with the maintainers. When submitting a pull request, provide a comprehensive description of your work and the testing you have conducted.

We are committed to fostering a collaborative and inclusive community where every contribution, whether it is a small typo fix or a major feature implementation, is valued. By contributing to Redway, you are helping to build a more efficient and accessible way to run Android in containerized environments. All contributions are released under the GPL-2.0 License, preserving the open-source nature of the project for everyone. We look forward to your pull requests and to working together to make Redway the premier choice for redroid orchestration.

## License

GPL-2.0 License

## Credits

- [redroid](https://github.com/remote-android/redroid-doc) - Remote-Android
  project
- [LXC](https://linuxcontainers.org/) - Linux Containers
- [skopeo](https://github.com/containers/skopeo) - OCI image operations
- [umoci](https://github.com/opencontainers/umoci) - OCI image unpacker

## Links

- Redroid Documentation: https://github.com/remote-android/redroid-doc
- LXC Documentation: https://linuxcontainers.org/lxc/documentation/
- LXC Getting Started: https://linuxcontainers.org/lxc/getting-started/
- LXC Configuration: https://linuxcontainers.org/lxc/manpages/man5/lxc.container.conf.5.html
- LXC Command Line: https://linuxcontainers.org/lxc/manpages/man1/
- OCI Image Spec: https://github.com/opencontainers/image-spec
