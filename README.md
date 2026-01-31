# Reddock - Redroid Docker Manager

Reddock is a lightweight Go-based **Docker container manager** for
[redroid](https://github.com/remote-android/redroid-doc). It provides a
streamlined CLI for managing Android containers using Docker, focusing on ease
of use, performance, and reliability.

## Key Features

- **Docker-Based** - Leverages Docker for robust container management
- **Simplified CLI** - easy-to-use commands for init, start, stop, and removal
- **Kernel Module Management** - Automatically checks and attempts to load
  required kernel modules (`binder_linux`)
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

## Installation

1. **Install Dependencies**

   You need to install go/golang to build, and then lzip, tar, and xz-utils on
   your Linux systems.

2. **Clone and Build**

   ```bash
   git clone https://github.com/jimed-rand/reddock.git
   cd reddock
   make build
   ```

3. **Install (Optional)**

   ```bash
   sudo make install
   ```

## Usage

> [!IMPORTANT]
> Most commands require `sudo` to manage container runtimes and kernel modules.

### 1. Initialize a Container

```bash
sudo reddock init my-android redroid/redroid:13.0.0-latest
```

### 2. Start the Container

```bash
sudo reddock start my-android
```

### 3. Connect via ADB

```bash
reddock adb-connect my-android
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
| `version`               | Show version information                            |

## Troubleshooting

- **Container must be running**: Some operations only work on active containers.

## Credits

- [redroid](https://github.com/remote-android/redroid-doc) - Remote-Android
  project.

## License

GPL-2.0 license
