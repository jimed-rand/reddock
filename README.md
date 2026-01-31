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

## Prerequisites

- **Docker** - Installed and running
- **Kernel Modules** - Linux kernel with `binder` and `ashmem` support (Reddock
  attempts to load these automatically)
- **ADB** - For Android debugging and connection management
- **Root Access** - Required for managing Docker and kernel modules

## Installation

### From Source

```bash
git clone https://github.com/jimed-rand/reddock.git
cd reddock
make build
sudo make install
```

### Manual Compilation

```bash
go build -o reddock .
sudo cp reddock /usr/local/bin/
sudo chmod +x /usr/local/bin/reddock
```

## Quick Start

### 1. Initialize Container

```bash
sudo reddock init android13 redroid/redroid:13.0.0_64only-latest
```

### 2. Start Container

```bash
sudo reddock start android13
```

### 3. Connect via ADB

```bash
sudo reddock adb-connect android13
```

## Usage

| Command               | Description                         |
| --------------------- | ----------------------------------- |
| `init <name> [image]` | Initialize a new Redroid container  |
| `start <name> [-v]`   | Start a container (use -v for logs) |
| `stop <name>`         | Stop a running container            |
| `restart <name> [-v]` | Restart a container                 |
| `status <name>`       | Show container status and info      |
| `shell <name>`        | Enter the container shell           |
| `adb-connect <name>`  | Connect to the container via ADB    |
| `log <name>`          | Show container logs                 |
| `list`                | List all Reddock-managed containers |
| `remove <name>`       | Remove a container and its data     |

## Configuration

Configuration is stored in `~/.config/reddock/config.json`. You can customize
GPU modes, ports, and data paths there.

### GPU Modes

- `guest`: Software rendering (most compatible)
- `host`: Hardware acceleration (requires compatible host GPU)
- `auto`: Auto-detect (recommended)

## Why Reddock?

Reddock replaces the previous LXC-based implementation (Redway) to provide a
more stable, widely compatible, and easier-to-manage environment for Redroid
using Docker's industry-standard containerization.

## License

GPL-2.0 license

## Credits

- [redroid](https://github.com/remote-android/redroid-doc) - Remote-Android
  project
- [Docker](https://www.docker.com/) - Container platform
