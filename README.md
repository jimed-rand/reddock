# Reddock - Redroid Docker Manager

Reddock is a lightweight Go-based **Docker container manager** for
[redroid](https://github.com/remote-android/redroid-doc). It provides a
streamlined CLI for managing Android containers using Docker, focusing on ease
of use, performance, and reliability.

## Key Features

- **Docker-Based** - Leverages Docker for robust container management
- **Simplified CLI** - easy-to-use commands for init, start, stop, and removal
- **Kernel Module Management** - Automatically checks and attempts to load
  required kernel modules (`binder_linux`, `ashmem_linux`)
- **ADB Integration** - Built-in ADB connection management with automatic port
  mapping
- **GPU Acceleration Support** - Easy configuration for GPU rendering modes
  (`host`, `guest`, `auto`)
- **Persistent Storage** - Android data persists across container restarts using
  Docker volumes

## What is Redroid?

[Redroid](https://github.com/remote-android/redroid-doc) (Remote Android) is a
GPU-accelerated **AIC (Android In Container)** solution that allows you to run
Android in containers with near-native performance. It is commonly used for
cloud gaming, automated testing, and virtual mobile infrastructure.

## Requirements

- **Linux OS**: Ubuntu 20.04+, Debian 11+, etc.
- **Container Runtime**: Docker OR Podman
- **Binder Modules**: `binder_linux` (usually available on modern kernels or can
  be loaded)

> [!NOTE]
> Ashmem is no longer required for Redroid on modern kernels.

## Installation

1. **Install Go** (if not installed):
   ```bash
   sudo apt install golang
   ```

2. **Clone and Build**:
   ```bash
   git clone https://github.com/jimed-rand/reddock.git
   cd reddock
   make build
   ```

3. **Install (Optional)**:
   ```bash
   sudo make install
   ```

## Usage

> [!IMPORTANT]
> Most commands require `sudo` to manage container runtimes and kernel modules.

You can run `reddock` with `docker` or `podman`. If your user has permissions to
use the container runtime (e.g., added to `docker` group), you do not need
`sudo`.

### 1. Initialize a Container

Downloads the image and prepares the environment.

```bash
sudo reddock init android13
```

### 2. Start the Container

Starts the container instance.

```bash
sudo reddock start android13
```

### 3. Connect via ADB

```bash
reddock adb-connect android13
```

## Commands

| Command                 | Description                                         |
| ----------------------- | --------------------------------------------------- |
| `init <name> [image]`   | Initialize a new Redroid container                  |
| `start <name> [-v]`     | Start a container (use -v for logs)                 |
| `stop <name>`           | Stop a running container                            |
| `restart <name> [-v]`   | Restart a container                                 |
| `status <name>`         | Show container status and info                      |
| `shell <name>`          | Enter the container shell                           |
| `adb-connect <name>`    | Connect to the container via ADB                    |
| `log <name>`            | Show container logs                                 |
| `list`                  | List all Reddock-managed containers                 |
| `remove <name> [--all]` | Remove a container, data, and optionally image (-a) |

## Configuration

Configuration is stored in `~/.config/reddock/config.json`. You can customize
GPU modes, ports, and data paths there.

### GPU Modes

- `guest`: Software rendering (most compatible)
- `host`: Hardware acceleration (requires compatible host GPU)
- `auto`: Auto-detect (recommended)

## Why Reddock?

Reddock provides a stable, widely compatible, and easy-to-manage environment for
Redroid using modern containerization standards (Docker/Podman). It streamlines
the setup process, handles kernel module requirements, and manages persistent
Android data effortlessly.

## License

GPL-2.0 license

## Credits

- [redroid](https://github.com/remote-android/redroid-doc) - Remote-Android
  project
- [Docker](https://www.docker.com/) - Container platform
