# terpodman

Podman for Termux. Run real OCI containers on any non-rooted Android phone.
WARN, it's still in beta, always use the one linear or the latest binarys!!! 
Also deleting, uninstalling the packages must be done manually for now. 

```
terpodman pull alpine
terpodman run -it alpine sh
```

That works on a stock Pixel, Galaxy, or any 2018+ Android phone with Termux installed. No root, no custom kernel, no `tsu`.

## What it is

A single 8.5 MB Go binary that gives you the `podman` CLI on Termux. It pulls images from Docker Hub, runs them inside `proot` (user-space syscall interception, no root needed), builds new images from Dockerfiles, and orchestrates multi-container apps with `docker-compose.yml`.

## Install

```bash
curl -fsSL https://github.com/jimkardy/terpodman/releases/latest/download/install.sh | bash
```

That's it. The script:

1. Detects your phone's CPU (arm64 or arm)
2. Downloads the right binary from the v0.1.0 GitHub release
3. Verifies the SHA-256 checksum
4. Installs to `$PREFIX/bin/terpodman`
5. Runs `pkg install proot` if proot is missing

Then verify it works:

```bash
terpodman --version
terpodman pull alpine
terpodman run -it alpine sh
```

## What you can do

```bash
# Images
terpodman pull nginx
terpodman images
terpodman tag nginx mynginx:1
terpodman rmi nginx

# Containers
terpodman run -d --name web -p 8080:80 nginx
terpodman ps
terpodman logs web
terpodman exec web nginx -s reload
terpodman stop web
terpodman start web
terpodman rm web

# Build from Dockerfile
terpodman build -t myapp:v1 .

# Volumes
terpodman volume create data
terpodman run -v data:/data alpine sh

# Pods (group containers)
terpodman pod create --name mypod
terpodman run -d --pod mypod nginx

# Compose
terpodman compose up -d
terpodman compose down
```

## What doesn't work (and never will, without root)

These are Android kernel limits, not bugs:

- `--memory` / `--cpus` — no cgroups on Android non-root. Flags are accepted and ignored.
- Real bridge networking — no network namespaces. Containers share the host network. Port forwarding (`-p`) works via a user-space proxy.
- Seccomp profiles — useless without PID namespaces.
- Overlayfs — no `mount()`. Each container gets a full copy of its image's rootfs.
- True `exec` into a running container — `terpodman exec` runs a fresh proot against the same rootfs. Shared filesystem, fresh environment.

`terpodman info` prints this list at runtime.

## Port forwarding (`-p`)

`terpodman run -p HOST:CONTAINER IMAGE` works for both foreground and detached
containers, but the mechanism differs because proot does not isolate
networking — the container's process binds directly to the host's loopback
interface.

### Foreground (`terpodman run -p 8080:80 nginx`)

A user-space TCP/UDP proxy is started **in the CLI process** on
`127.0.0.1:8080`. It forwards bytes to `127.0.0.1:80`, where the container's
process is listening. The proxy dies with the CLI when the container exits —
no orphan processes.

### Detached (`terpodman run -d -p 8080:80 nginx`)

Because the CLI exits immediately after starting a detached container, an
in-process proxy would die with it. Instead, terpodman spawns a **sidecar
process** (`terpodman __internal-proxy`) for each published port:

- The sidecar runs in its own process group (`Setpgid`) so it survives the
  CLI's exit and is not killed by signals sent to the parent's group.
- The sidecar polls the container's metadata.json once per second. When the
  container transitions out of the "running" state (because `terpodman stop`
  was called, the container exited naturally, or `terpodman rm` removed it),
  the sidecar stops the proxy and exits cleanly.
- A startup error (e.g. host port already in use) is written to the sidecar's
  stderr. The parent CLI reads stderr with a 1-second timeout and surfaces
  the error to the user as `warning: port forwarding failed: ...` before
  returning the container ID.

To see the sidecar processes for a detached container:

```bash
ps -ef | grep __internal-proxy
```

To force-kill a leaked sidecar (rare; should only happen if the container's
metadata was deleted out from under it):

```bash
pkill -f __internal-proxy
```

### Limitations

- The host port is bound to `127.0.0.1` only (not `0.0.0.0`) for safety on
  non-root Termux. Use `adb forward` or `socat` to expose a port externally.
- UDP forwarding is best-effort: per-client upstream sockets are never
  evicted on an idle timeout, so a long-lived sidecar will accumulate one
  goroutine per unique client address.
- If the host port is already in use (by another container or by an
  unrelated process), the bind fails. The foreground path logs a warning;
  the detached path surfaces the error before returning the container ID.

## Security

Containers are isolated by proot (ptrace-based). This is **not** as strong as real podman isolation. A malicious process running as root inside a container can theoretically escape by ptracing the host proot process. **Do not run untrusted images.** Same security model as `proot-distro`.

Mitigations: all state lives under `~/.local/share/terpodman/` owned by your Termux user. Port forwarding binds to `127.0.0.1` only.

## Build from source

Needs Go 1.21+ (`pkg install golang` in Termux):

```bash
git clone https://github.com/jimkardy/terpodman
cd terpodman
make build       # current host
make all         # arm64 + arm
```

## License

Apache 2.0.
