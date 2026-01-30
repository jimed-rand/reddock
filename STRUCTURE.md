# Redway Project Structure

This document describes the organization of the Redway project source code.

> [!WARNING]
> **Technical Documentation Notice:** This information is strictly reserved for developers, system administrators, and technical maintenance personnel. The contents of this document detail the internal structural organization and architectural logic of the Redway project, which are not relevant to the typical end-user experience. It is imperative to understand that the directory layouts, build scripts, and module definitions described below are critical components of the software's infrastructure.

Misinterpretation or unauthorized modification of these structural elements by non-technical users could result in significant system instability, broken dependencies, or complete failure of the build environment. This document serves as a comprehensive roadmap for those responsible for the ongoing development, debugging, and deployment of the application. It provides the necessary context for navigating the codebase efficiently and maintaining the integrity of the project's design patterns. End-users seeking assistance with application functionality should refer to the official user manual rather than this internal technical specification.

## Directory Layout

redway/
├── main.go                  # Application entry point
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums (generated)
├── Makefile                 # Build automation
│
├── cmd/                     # Command-line interface
│   ├── command.go           # Command dispatcher and routing
│   └── usage.go             # Help text and usage information
│
├── pkg/                     # Core application packages
│   ├── config/              # Configuration management
│   │   └── config.go        # JSON config read/write, defaults
│   │
│   ├── container/           # Container lifecycle management
│   │   ├── initializer.go   # Container initialization workflow
│   │   └── manager.go       # Start/stop/restart/remove operations
│   │
│   └── utils/               # Utility functions and helpers
│       ├── adb.go           # ADB connection information
│       ├── shell.go         # Container shell access
│       ├── status.go        # Status reporting
│       └── log.go           # Log file viewing

## Package Descriptions

### main.go

Entry point for the application. Handles:
- Command-line argument parsing
- Routing to appropriate command handlers
- Top-level error handling

**Key Functions:**
- `main()` - Program entry point

### cmd/

Command-line interface implementation.

**command.go**
- Command struct and dispatcher
- Routes commands to appropriate handlers
- Manages command execution flow

**usage.go**
- Help text and usage information
- Command descriptions
- Example usage

**Key Functions:**
- `NewCommand()` - Create command instance
- `Execute()` - Execute command
- `PrintUsage()` - Display help

### pkg/config/

Configuration management package.

**config.go**
- JSON configuration structure
- Load/save configuration
- Default values
- Path helpers

**Key Types:**
```go
type Config struct {
    ContainerName string
    ImageURL      string
    DataPath      string
    LogFile       string
    GPUMode       string
    Initialized   bool
}
```

**Key Functions:**
- `Load()` - Load configuration from disk
- `Save()` - Save configuration to disk
- `GetDefault()` - Get default configuration
- `GetConfigPath()` - Get config file path

### pkg/container/

Container lifecycle management.

**initializer.go**
- Container initialization workflow
- LXC OCI container creation
- Configuration adjustments
- Prerequisite checking

**Steps:**
1. Check kernel modules (binder - advisory, supports binderfs)
2. Check LXC tools
3. Check LXC networking
4. Adjust OCI template
5. Check required tools (skopeo, umoci, jq)
6. Create LXC container from OCI image
7. Create data directory
8. Adjust container configuration
9. Apply networking workaround

**manager.go**
- Container start/stop/restart
- Container status checking
- Container removal
- Information retrieval

**Key Types:**
```go
type Initializer struct {
    config *config.Config
    image  string
}

type Manager struct {
    config *config.Config
}

type Lister struct{}
```

**Key Functions:**
- `Initialize()` - Initialize new container
- `Start()` - Start container
- `Stop()` - Stop container
- `Restart()` - Restart container
- `Remove()` - Remove container
- `IsRunning()` - Check if running
- `GetInfo()` - Get container info
- `GetIP()` - Get container IP
- `GetPID()` - Get container PID

### pkg/utils/

Utility functions and helpers.

**adb.go**
- ADB connection information display
- Connection instructions

**shell.go**
- Container shell access via nsenter
- Interactive shell session

**status.go**
- Comprehensive status reporting
- Container information display
- Configuration display

**log.go**
- Log file monitoring
- Tail -f functionality
- Signal handling

**Key Types:**
```go
type ShellManager struct {
    manager *container.Manager
}

type AdbManager struct {
    manager *container.Manager
}

type StatusManager struct {
    manager *container.Manager
    config  *config.Config
}

type LogManager struct {
    config *config.Config
}
```

## Data Flow

### Initialization Flow

```
main.go
  └─> cmd.Execute()
       └─> container.NewInitializer()
            └─> initializer.Initialize()
                 ├─> checkKernelModules()
                 ├─> checkLXCTools()
                 ├─> checkLXCNetworking()
                 ├─> adjustOCITemplate()
                 ├─> checkRequiredTools()
                 ├─> createContainer()
                 ├─> createDataDirectory()
                 ├─> adjustContainerConfig()
                 └─> applyNetworkingWorkaround()
```

### Start Flow

```
main.go
  └─> cmd.Execute()
       └─> container.NewManager()
            └─> manager.Start()
                 ├─> IsRunning() check
                 └─> lxc-start execution
```

### Status Flow

```
main.go
  └─> cmd.Execute()
       └─> utils.NewStatusManager()
            └─> status.Show()
                 ├─> config.Load()
                 ├─> manager.GetInfo()
                 ├─> manager.GetIP()
                 └─> manager.GetPID()
```

## Configuration Files

### Runtime Configuration

**Location:** `~/.config/redway/config.json`

**Structure:**
```json
{
  "container_name": "redroid",
  "image_url": "docker://redroid/redroid:13.0.0_64only-latest",
  "data_path": "/home/user/data-redroid",
  "log_file": "redroid.log",
  "gpu_mode": "guest",
  "initialized": true
}
```

### LXC Container Configuration

**Location:** `/var/lib/lxc/{container_name}/config`

**Generated by:** `initializer.adjustContainerConfig()`

**Key Settings:**
- Init command with redroid parameters
- Apparmor profile (unconfined)
- Autodev settings
- Data directory mount

## External Dependencies

### Go Packages

- Standard library only
- No external Go dependencies

### System Tools

Required at runtime:
- `lxc-create` - Container creation
- `lxc-start` - Container startup
- `lxc-stop` - Container shutdown
- `lxc-info` - Container information
- `lxc-destroy` - Container removal
- `skopeo` - OCI image handling
- `umoci` - OCI image unpacking
- `jq` - JSON processing
- `nsenter` - Namespace entry
- `tail` - Log viewing

## Build Process

### Standard Build

```bash
make build
```

Produces: `./redway`

### Static Build

```bash
make static
```

Produces: `./redway` (static binary)

### Installation

```bash
sudo make install
```

Installs to: `/usr/local/bin/redway`

## Testing Structure

Currently no automated tests. Future structure:

```
redway/
├── pkg/
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go
│   ├── container/
│   │   ├── initializer.go
│   │   ├── initializer_test.go
│   │   ├── manager.go
│   │   └── manager_test.go
│   └── utils/
│       ├── adb_test.go
│       └── status_test.go
```

## Design Decisions

### Why This Structure?

1. **Separation of Concerns**
   - CLI layer (`cmd/`)
   - Business logic (`pkg/`)
   - Clear boundaries

2. **Package Organization**
   - `config` - Pure configuration
   - `container` - Container operations
   - `utils` - Helper functions

3. **Cross-Distro Compatibility**
   - Advisory kernel checks (warn, don't fail)
   - Supports multiple binder implementations (module, built-in, binderfs)
   - No kernel module loading (user controls system configuration)
   - Works on Ubuntu, Debian, Fedora, Arch, openSUSE, Gentoo, and more

4. **Minimal Dependencies**
   - Standard library only
   - Relies on system tools
   - Easy to audit and maintain

### Future Expansion

Easy to add:
- `pkg/network/` - Network management
- `pkg/gpu/` - GPU configuration
- `pkg/backup/` - Backup/restore
- `pkg/monitor/` - Resource monitoring

## Code Conventions

### File Naming

- One primary type per file
- File named after main type
- Example: `Manager` type in `manager.go`

### Package Naming

- Lowercase single words
- Descriptive and concise
- No underscores or mixed case

### Error Handling

- Always return errors, never panic
- Wrap errors with context
- User-friendly error messages

### Command Output

- Errors go to stderr
- Info messages to stdout

## Modification Guidelines

### Adding a New Command

1. Add command in `cmd/command.go`:
   ```go
   case "newcmd":
       return c.executeNewCmd()
   ```

2. Implement handler:
   ```go
   func (c *Command) executeNewCmd() error {
       // Implementation
   }
   ```

3. Update `cmd/usage.go`:
   ```go
   fmt.Println("  newcmd      Description")
   ```

### Adding a New Package

1. Create directory: `pkg/newpkg/`
2. Create files with package declaration: `package newpkg`
3. Import in relevant files: `import "redway/pkg/newpkg"`
4. Update documentation

### Modifying Configuration

1. Update struct in `pkg/config/config.go`
2. Update `GetDefault()` function
3. Update documentation
4. Consider migration for existing configs

## Integration Points

### LXC Integration

- Uses LXC command-line tools
- No library dependencies
- Parses text output
- Platform independent

### OCI Registry Integration

- Via `skopeo` and `umoci`
- OCI-compliant image handling
- Supports any OCI-compatible registry (including Docker Hub)
- No Docker daemon required

### ADB Integration

- Provides connection information
- User installs ADB separately
- No direct ADB calls

## Performance Considerations

### Fast Operations

- Container start/stop
- Status checking
- Configuration loading

### Slow Operations

- Container initialization (downloads image)
- Container removal (destroys filesystem)

### Optimization Opportunities

- Cache LXC info calls
- Parallel tool checking
- Concurrent container operations

## Security Considerations

### Root Requirements

- LXC operations require root
- Sudo used for privileged operations
- User data in user's home directory

### Configuration Security

- Config stored in user directory
- No passwords or secrets
- World-readable by design

### Container Security

- Apparmor unconfined (required for Android)
- Binder devices exposed
- Network accessible

## Maintenance Notes

### Regular Updates

- Keep Go version current
- Monitor LXC changes
- Track redroid updates
- Update documentation

### Compatibility

- Test on major distros
- Check kernel requirements
- Verify tool availability

### Documentation

- Keep README current
- Update examples
- Document breaking changes
- Maintain this structure guide
