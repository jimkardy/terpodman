# Terpodman Implementation Report

## Executive Summary

This report documents the implementation of **terpodman** - a Podman-compatible CLI tool designed specifically for Termux on Android devices without root access or kernel support. The implementation is based on analysis of the official Podman repository and adapted for the constraints of the Android/Termux environment.

## Current Status

**Version:** 0.1.0-beta  
**Binary Size:** 2.6 MB (stripped)  
**Source Lines:** 791 lines of Go code  
**Build Target:** Android ARM64/ARM (CGO disabled)

## Architecture Analysis from Official Podman Repository

### Key Components Studied

1. **Command Structure** (`cmd/podman/`)
   - Main entry point: `main.go`
   - Root command setup: `root.go`
   - Container commands: `containers/run.go`, `containers/create.go`
   - Image management: `images/`
   - Pod management: `pods/`

2. **Core Runtime** (`libpod/`)
   - Runtime initialization: `runtime.go`
   - Container lifecycle: `container.go`, `container_internal.go`
   - Container execution: `container_exec.go`
   - State management: Various state backends

3. **Key Dependencies Identified**
   - `github.com/containers/image` - Image pulling and management
   - `github.com/opencontainers/runtime-spec` - OCI specification
   - `github.com/sirupsen/logrus` - Logging
   - `github.com/spf13/cobra` - CLI framework
   - `go.podman.io/storage` - Storage backend

### What Cannot Be Used in Termux (No Root/Kernel Support)

The following Podman features are **impossible** without root access:

| Feature | Reason | Podman Dependency |
|---------|--------|-------------------|
| Cgroups (--memory, --cpus) | No cgroup filesystem access | `pkg/cgroups` |
| Network namespaces | Requires CAP_NET_ADMIN | `libnetwork/netns` |
| PID namespaces | Requires kernel support | `pkg/rootless` |
| Mount namespaces | Requires CAP_SYS_ADMIN | `pkg/specgen` |
| OverlayFS | No mount() syscall | `drivers/overlay` |
| Seccomp profiles | Requires ptrace capabilities | `pkg/seccomp` |
| Real bridge networking | No network namespace isolation | `libnetwork` |
| True container exec | Shares no namespaces | `container_exec.go` |

## Terpodman Implementation Details

### Data Structures

```go
type Image struct {
    ID, Name, Tag string
    Size int64
    Created time.Time
    RootFSPath, MetaPath string
}

type Container struct {
    ID, Name, Image string
    State string // created, running, stopped, exited
    Command []string
    Ports []PortMapping
    Volumes []VolumeMount
    ProotPID int
    LogPath string
}

type PortMapping struct {
    HostIP string
    HostPort, ContainerPort int
    Protocol string // tcp, udp
    ProxyPID int
}

type Volume struct {
    Name, Driver, Mountpoint string
    CreatedAt time.Time
}

type Pod struct {
    ID, Name, State string
    Containers []string // container IDs
}
```

### Directory Structure

```
~/.local/share/terpodman/
├── images/      # Image metadata JSON files
├── containers/  # Container metadata JSON files
├── volumes/     # Volume directories
└── pods/        # Pod metadata JSON files
```

### Implemented Commands

| Command | Status | Notes |
|---------|--------|-------|
| `pull` | Skeleton | Parses image:tag, checks proot availability |
| `run` | Skeleton | Parses all standard flags (-d, -it, -v, -p, -e, etc.) |
| `ps` | Working | Lists containers from metadata files |
| `images` | Working | Lists images from metadata files |
| `stop` | Skeleton | Needs proot process management |
| `start` | Skeleton | Needs proot process management |
| `rm` | Skeleton | Needs metadata + rootfs cleanup |
| `rmi` | Skeleton | Needs image cleanup |
| `exec` | Skeleton | Runs fresh proot (not true exec) |
| `logs` | Working | Reads container log files |
| `build` | Skeleton | Needs Dockerfile parser |
| `volume create/ls/rm/inspect` | Working | Full implementation |
| `pod create/ls/rm/inspect` | Skeleton | Metadata management |
| `compose up/down/ps` | Skeleton | Needs YAML parser |
| `info` | Working | Shows limitations and architecture |

### Flag Parsing (from Podman's createFlags)

The implementation parses these standard Podman flags:
- `--name` - Container name
- `-d, --detach` - Run in background
- `-i, --interactive` - Keep stdin open
- `-t, --tty` - Allocate pseudo-TTY
- `-v, --volume` - Mount volumes
- `-p, --publish` - Port mappings
- `-e, --env` - Environment variables
- `-w, --workdir` - Working directory
- `-u, --user` - User inside container
- `--rm` - Auto-remove on exit

## Required Next Steps for Full Implementation

### 1. Image Pulling (Highest Priority)

**What's needed:**
- Integrate `github.com/containers/image/v5` for OCI image pulling
- Download and extract layers to user-space
- Store layer metadata in JSON format
- Handle Docker Hub authentication

**Reference from Podman:**
- `libimage/manifest.go`
- `libimage/image.go`
- `pkg/api/handlers/libpod/images.go`

### 2. Proot Integration

**What's needed:**
- Detect proot binary (`exec.LookPath("proot")`)
- Build proot command line with proper bindings
- Manage proot subprocess lifecycle
- Handle stdin/stdout/stderr forwarding
- Implement detach mode with PID file

**Proot command structure:**
```bash
proot \
  -r /path/to/container/rootfs \
  -b /host/path:/container/path \
  -w /working/dir \
  /bin/sh -c "command"
```

### 3. Container Lifecycle Management

**What's needed:**
- `containerInternalRun()` - Create and start container
- `containerInternalStop()` - Send signal to proot, wait for exit
- `containerInternalStart()` - Restart stopped container
- State transitions: created → running → stopped/exited

**Reference from Podman:**
- `libpod/container_internal.go`
- `libpod/container_api.go`

### 4. Port Forwarding Proxy

**What's needed:**
- User-space TCP/UDP proxy (as described in readme)
- Sidecar process for detached containers
- Poll container metadata for lifecycle
- Bind to 127.0.0.1 only (security)

**Reference from Podman:**
- `cmd/podman/rootlessport/`
- Readme section on port forwarding

### 5. Dockerfile Build Support

**What's needed:**
- Parse Dockerfile instructions
- Execute each instruction in proot environment
- Create layers as tar archives
- Generate image metadata

**Reference from Podman:**
- `buildah/pkg/dockerfile`
- `pkg/bindings/images/build.go`

### 6. Compose Support

**What's needed:**
- Parse docker-compose.yml (use `gopkg.in/yaml.v3`)
- Create pod for service grouping
- Start containers in dependency order
- Handle volume and network definitions

**Reference from Podman:**
- `cmd/podman/compose.go`
- `pkg/domain/entities/compose.go`

## Security Considerations

1. **Ptrace-based isolation** - proot uses ptrace, which is weaker than namespaces
2. **Localhost-only binding** - Port forwarding binds to 127.0.0.1
3. **User-owned storage** - All data in `~/.local/share/terpodman/`
4. **No setuid binaries** - Everything runs as the Termux user

## Testing Performed

```bash
# Version check
./terpodman --version
# Output: terpodman version 0.1.0-beta

# Info command
./terpodman info
# Output: Shows limitations and architecture

# Volume management
./terpodman volume create testvol
./terpodman volume ls
# Output: DRIVER    VOLUME NAME
#         local     testvol

# Run command parsing
./terpodman run -it --name test alpine sh
# Output: Shows parsed options correctly

# Ps command (empty - no containers yet)
./terpodman ps -a
# Output: Header row only
```

## Build Instructions

### For Current Host
```bash
cd /workspace
make build
# Or: CGO_ENABLED=0 go build -ldflags="-s -w" -o terpodman .
```

### For ARM64 (most Termux devices)
```bash
make build-arm64
# Or: CGO_ENABLED=0 GOOS=android GOARCH=arm64 go build -ldflags="-s -w" -o terpodman-arm64 .
```

### For ARM (32-bit)
```bash
make build-arm
# Or: CGO_ENABLED=0 GOOS=android GOARCH=arm go build -ldflags="-s -w" -o terpodman-arm .
```

### Install to Termux
```bash
make install
# Copies to $PREFIX/bin or ~/bin
```

## Files Created/Modified

| File | Purpose | Lines |
|------|---------|-------|
| `main.go` | Core implementation | 791 |
| `go.mod` | Go module definition | 12 |
| `Makefile` | Build automation | 22 |
| `terpodman` | Compiled binary | 2.6MB |

## Comparison: Podman vs Terpodman

| Aspect | Podman | Terpodman |
|--------|--------|-----------|
| Root required | Optional (rootless mode) | No |
| Kernel namespaces | Yes | No |
| Isolation mechanism | Namespaces + cgroups | proot (ptrace) |
| Network isolation | Yes (netns) | No (host network) |
| Resource limits | Yes (cgroups) | No |
| OverlayFS | Yes | No (full copy) |
| Binary size | ~50MB | ~2.6MB |
| Dependencies | systemd, runc, conmon | proot only |
| Security model | Strong (kernel-enforced) | Weaker (userspace) |

## Conclusion

This implementation provides a **solid foundation** for terpodman with:

✅ Complete CLI structure matching Podman's interface  
✅ All command handlers implemented (skeleton to working)  
✅ Proper flag parsing from Podman patterns  
✅ Data structures for images, containers, volumes, pods  
✅ Metadata storage system  
✅ Volume management fully functional  
✅ Clear documentation of limitations  

**Remaining work** focuses on:
1. Integrating containers/image library for actual image pulling
2. Proot subprocess management for container execution
3. Port forwarding proxy implementation
4. Dockerfile parsing and build support
5. Compose YAML parsing and orchestration

The architecture follows Podman's design patterns while respecting Termux's constraints. The codebase is ready for incremental feature addition.

---

*Report generated: $(date)*  
*Based on analysis of github.com/containers/podman (commit: latest)*
