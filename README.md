# Reddock

**Reddock** is a small Go CLI for managing [redroid](https://github.com/remote-android/redroid-doc) containers with Docker on **Linux x86_64**. It focuses on a straightforward workflow: initialize instances, start/stop, ADB, logs, and optional GPU settings—without a heavy stack around it.

## Features

- **Docker-first** — Uses the Docker API for container lifecycle and volumes.
- **Simple commands** — `init`, `start`, `stop`, `restart`, `status`, `shell`, `list`, `remove`, and more.
- **Kernel modules** — Detects `binder_linux` (loaded, sysfs, `modinfo`, or `.ko` under `/lib/modules/...` for DKMS/KMP), the distinct `binder` module when present, legacy `/dev/binder*`, binderfs layout `/dev/binderfs/*`, and `/proc/filesystems` binderfs support; then best-effort `modprobe binder_linux` when appropriate.
- **ADB** — Helpers to connect to the emulated device over the published port.
- **GPU modes** — Configure rendering (`host`, `guest`, `auto`) when supported by your setup.
- **Persistent data** — Android user data can live in Docker volumes across restarts.

## Requirements

| Requirement | Notes |
| ----------- | ----- |
| OS | **Linux**, **x86_64** (amd64) |
| Docker | Installed and usable by your user (often `docker` group or `sudo`) |
| Permissions | Many commands need elevated privileges for Docker and kernel modules |

Optional for building from source: **Go 1.21+**, `make`, `tar`, **xz** (for `make dist-pack`).

## Installation

### Install script (recommended)

Installs the **`reddock` binary** from GitHub Releases into `$PREFIX/bin` (default `/usr/local/bin`).

```bash
curl -fsSL https://raw.githubusercontent.com/jimed-rand/reddock/main/scripts/install.sh | sh
```

With options:

```bash
curl -fsSL https://raw.githubusercontent.com/jimed-rand/reddock/main/scripts/install.sh | sh -s -- --tag v0.1.0 --prefix /usr/local
```

| Variable / flag | Meaning |
| --------------- | ------- |
| `REDDOCK_REPO` | `owner/repo` on GitHub (default: `jimed-rand/reddock`) |
| `PREFIX` | Install prefix; binary at `$PREFIX/bin/reddock` |
| `--tag` | Exact release tag (default: **latest** release) |
| `--dry-run` | Show URL and paths only |

### Manual download

From the [Releases](https://github.com/jimed-rand/reddock/releases) page, download for your tag:

- **`reddock`** — standalone executable (chmod +x, place on `PATH`), or  
- **`reddock-<version>-linux-amd64.tar.xz`** — archive containing the binary plus `README.md` and `LICENSE`.

Example:

```bash
tar -xJf reddock-v0.1.0-linux-amd64.tar.xz
sudo install -Dm755 reddock /usr/local/bin/reddock
```

### Build from source

```bash
git clone https://github.com/jimed-rand/reddock.git
cd reddock
make static
sudo make install PREFIX=/usr/local
```

`make install` builds a dynamically linked binary and installs it; for a static binary, run `make static` then copy `./reddock` yourself or use `install` on that file.

## Versioning and releases

GitHub **tags** (`vMAJOR.MINOR.PATCH`) define what is shipped; the binary always reports two strings:

| Field | Meaning |
| ----- | ------- |
| **Release** | The public distribution tag (e.g. `v2.4.2`) for builds produced for a GitHub Release. Empty for ad‑hoc local builds. |
| **Snapshot** | Internal build fingerprint: short git commit SHA plus `-dirty` when the tree is not clean. |

Run `reddock version` to see both. `reddock --help` shows the release tag in the banner when present, otherwise the snapshot.

| How | What happens |
| --- | ------------ |
| **Local `make`** | Embeds **Snapshot** from git; **Release** unset unless you set `RELEASE=…` (e.g. `make static RELEASE=v0.2.0`). |
| **Push tag `v*` / Actions workflow** | CI sets **Release** to the tag and **Snapshot** to the commit used for that build. |

To release from your machine without the Actions button: `git tag -a v0.2.0 -m "v0.2.0"`, then `git push origin v0.2.0`.

## Usage

> **Note:** Most operations need `sudo` (or root) when Docker or kernel modules require it.

### Initialize a container

```bash
sudo reddock init my-android redroid/redroid:13.0.0-latest
```

### Start and use ADB

```bash
sudo reddock start my-android
reddock adb-connect my-android
```

### CLI reference

| Command | Description |
| ------- | ----------- |
| `init <name> [image]` | Create a new Reddock-managed container |
| `start <name> [-v]` | Start (optional verbose logs) |
| `stop <name>` | Stop |
| `restart <name> [-v]` | Restart |
| `status <name>` | Status and info |
| `shell <name>` | Shell into the container |
| `adb-connect <name>` | Connect ADB to the instance |
| `log <name>` | Container logs |
| `list` | List Reddock-managed containers |
| `remove <name>` (`--image` / `-i`) | Remove container/data; optional image removal |
| `version` | Print Reddock version string |

Use `reddock --help` for the full flag list.

## Troubleshooting

- **Binder / binderfs** — `reddock status` shows host binder detection. Nodes may be `/dev/binder` (legacy) or `/dev/binderfs/binder` (binderfs). A packaged `binder_linux` (DKMS/KMP) is detected even before load via `modinfo` or a matching `.ko` under `/lib/modules/$(uname -r)/`.
- **Container not running** — Commands like `adb-connect` need a started container (`reddock start …`).
- **Docker permission denied** — Use `sudo` or add your user to the `docker` group and re-login.
- **Wrong architecture** — Prebuilt release binaries are **linux/amd64** only.

## Credits

- [redroid / Remote Android](https://github.com/remote-android/redroid-doc)

## License

GPL-2.0 — see [LICENSE](LICENSE).
