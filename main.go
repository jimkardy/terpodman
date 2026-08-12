package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
terpodmanHome = "TERPODMAN_HOME"
defaultHome   = "~/.local/share/terpodman"
imagesDir     = "images"
containersDir = "containers"
volumesDir    = "volumes"
podsDir       = "pods"
)

type Image struct {
ID         string    `json:"id"`
Name       string    `json:"name"`
Tag        string    `json:"tag"`
Size       int64     `json:"size"`
Created    time.Time `json:"created"`
Digest     string    `json:"digest,omitempty"`
RootFSPath string    `json:"rootfs_path"`
MetaPath   string    `json:"meta_path"`
}

type Container struct {
ID           string            `json:"id"`
Name         string            `json:"name"`
Image        string            `json:"image"`
ImageID      string            `json:"image_id"`
State        string            `json:"state"` // created, running, stopped, exited
Status       string            `json:"status"`
Created      time.Time         `json:"created"`
Started      *time.Time        `json:"started,omitempty"`
ExitedAt     *time.Time        `json:"exited_at,omitempty"`
ExitCode     int               `json:"exit_code"`
Command      []string          `json:"command"`
Env          []string          `json:"env"`
WorkingDir   string            `json:"working_dir"`
User         string            `json:"user"`
Ports        []PortMapping     `json:"ports"`
Volumes      []VolumeMount     `json:"volumes"`
Pod          string            `json:"pod,omitempty"`
RootFSPath   string            `json:"rootfs_path"`
MetaPath     string            `json:"meta_path"`
LogPath      string            `json:"log_path"`
ProotPID     int               `json:"proot_pid,omitempty"`
AttachKey    string            `json:"attach_key"`
}

type PortMapping struct {
HostIP        string `json:"host_ip"`
HostPort      int    `json:"host_port"`
ContainerPort int    `json:"container_port"`
Protocol      string `json:"protocol"` // tcp, udp
ProxyPID      int    `json:"proxy_pid,omitempty"`
}

type VolumeMount struct {
Source      string `json:"source"`
Destination string `json:"destination"`
ReadOnly    bool   `json:"read_only"`
}

type Volume struct {
Name       string `json:"name"`
Driver     string `json:"driver"`
Mountpoint string `json:"mountpoint"`
CreatedAt  time.Time `json:"created_at"`
}

type Pod struct {
ID         string            `json:"id"`
Name       string            `json:"name"`
State      string            `json:"state"`
Containers []string          `json:"containers"` // container IDs
Created    time.Time         `json:"created"`
MetaPath   string            `json:"meta_path"`
}

func main() {
if len(os.Args) < 2 {
printUsage()
os.Exit(1)
}

cmd := os.Args[1]
args := os.Args[2:]

switch cmd {
case "pull":
handlePull(args)
case "run":
handleRun(args)
case "ps":
handlePs(args)
case "images":
handleImages(args)
case "stop":
handleStop(args)
case "start":
handleStart(args)
case "rm":
handleRm(args)
case "rmi":
handleRmi(args)
case "exec":
handleExec(args)
case "logs":
handleLogs(args)
case "build":
handleBuild(args)
case "volume":
handleVolume(args)
case "pod":
handlePod(args)
case "compose":
handleCompose(args)
case "info":
handleInfo(args)
case "--version", "-v":
fmt.Println("terpodman version 0.1.0-beta")
case "help", "--help", "-h":
printUsage()
default:
fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n", cmd)
printUsage()
os.Exit(1)
}
}

func getTerpodmanHome() string {
if home := os.Getenv(terpodmanHome); home != "" {
return home
}
homeDir, _ := os.UserHomeDir()
return filepath.Join(homeDir, ".local", "share", "terpodman")
}

func ensureDirs() error {
home := getTerpodmanHome()
dirs := []string{
filepath.Join(home, imagesDir),
filepath.Join(home, containersDir),
filepath.Join(home, volumesDir),
filepath.Join(home, podsDir),
}
for _, dir := range dirs {
if err := os.MkdirAll(dir, 0755); err != nil {
return fmt.Errorf("failed to create directory %s: %w", dir, err)
}
}
return nil
}

func printUsage() {
fmt.Println(`terpodman - Podman for Termux (no root, no kernel support)

Usage:
  terpodman <command> [options]

Commands:
  pull <image>              Pull an image from Docker Hub
  run [options] <image>     Run a container
  ps                        List containers
  images                    List images
  stop <container>          Stop a running container
  start <container>         Start a stopped container
  rm <container>            Remove a container
  rmi <image>               Remove an image
  exec <container> <cmd>    Execute a command in a container
  logs <container>          Show container logs
  build -t <name> <path>    Build an image from Dockerfile
  volume <subcommand>       Manage volumes
  pod <subcommand>          Manage pods
  compose <subcommand>      Manage compose projects
  info                      Show system information
  --version, -v             Show version`)
}

// handlePull pulls an image from Docker Hub
func handlePull(args []string) {
if len(args) < 1 {
fmt.Fprintln(os.Stderr, "Error: image name required")
fmt.Fprintln(os.Stderr, "Usage: terpodman pull <image>[:tag]")
os.Exit(1)
}

imageName := args[0]
tag := "latest"

parts := strings.Split(imageName, ":")
if len(parts) > 1 {
imageName = parts[0]
tag = parts[1]
}

if err := ensureDirs(); err != nil {
fmt.Fprintf(os.Stderr, "Error: %v\n", err)
os.Exit(1)
}

fmt.Printf("Pulling %s:%s...\n", imageName, tag)

// Check if proot is available
prootPath, err := exec.LookPath("proot")
if err != nil {
fmt.Fprintln(os.Stderr, "Error: proot not found. Please install it with: pkg install proot")
os.Exit(1)
}

fmt.Printf("Using proot: %s\n", prootPath)
fmt.Println("Note: Full image pulling requires additional implementation.")
fmt.Println("This is a skeleton implementation based on podman architecture.")
}

// handleRun runs a container
func handleRun(args []string) {
if len(args) < 1 {
fmt.Fprintln(os.Stderr, "Error: image name required")
fmt.Fprintln(os.Stderr, "Usage: terpodman run [options] <image> [command]")
os.Exit(1)
}

if err := ensureDirs(); err != nil {
fmt.Fprintf(os.Stderr, "Error: %v\n", err)
os.Exit(1)
}

// Parse options - similar to podman's createFlags
var (
name       string
detach     bool
interactive bool
tty        bool
volumes    []string
ports      []string
env        []string
workdir    string
user       string
remove     bool
)

imageIdx := 0
for i, arg := range args {
switch arg {
case "--name":
if i+1 < len(args) {
name = args[i+1]
imageIdx = i + 2
}
case "-d", "--detach":
detach = true
case "-i", "--interactive":
interactive = true
case "-t", "--tty":
tty = true
case "-v", "--volume":
if i+1 < len(args) {
volumes = append(volumes, args[i+1])
}
case "-p", "--publish":
if i+1 < len(args) {
ports = append(ports, args[i+1])
}
case "-e", "--env":
if i+1 < len(args) {
env = append(env, args[i+1])
}
case "-w", "--workdir":
if i+1 < len(args) {
workdir = args[i+1]
}
case "-u", "--user":
if i+1 < len(args) {
user = args[i+1]
}
case "--rm":
remove = true
}
}

if imageIdx >= len(args) {
fmt.Fprintln(os.Stderr, "Error: image name required")
os.Exit(1)
}

imageName := args[imageIdx]
command := []string{}
if imageIdx+1 < len(args) {
command = args[imageIdx+1:]
}

fmt.Printf("Running container from %s\n", imageName)
if name != "" {
fmt.Printf("Container name: %s\n", name)
}
if detach {
fmt.Println("Running in detached mode")
}
if interactive {
fmt.Println("Interactive mode enabled")
}
if tty {
fmt.Println("TTY enabled")
}
for _, v := range volumes {
fmt.Printf("Volume mount: %s\n", v)
}
for _, p := range ports {
fmt.Printf("Port mapping: %s\n", p)
}
for _, e := range env {
fmt.Printf("Environment: %s\n", e)
}
if workdir != "" {
fmt.Printf("Working directory: %s\n", workdir)
}
if user != "" {
fmt.Printf("User: %s\n", user)
}
if remove {
fmt.Println("Auto-remove enabled")
}
if len(command) > 0 {
fmt.Printf("Command: %s\n", strings.Join(command, " "))
}

fmt.Println("\nNote: This is a skeleton implementation.")
fmt.Println("Full container execution requires proot integration.")
}

// handlePs lists containers
func handlePs(args []string) {
var allContainers bool

for _, arg := range args {
if arg == "-a" || arg == "--all" {
allContainers = true
}
}

fmt.Println("CONTAINER ID   IMAGE   COMMAND   CREATED   STATUS   PORTS   NAMES")

home := getTerpodmanHome()
containersPath := filepath.Join(home, containersDir)

entries, err := os.ReadDir(containersPath)
if err != nil {
return // No containers yet
}

for _, entry := range entries {
if !strings.HasSuffix(entry.Name(), ".json") {
continue
}

metaPath := filepath.Join(containersPath, entry.Name())
data, err := os.ReadFile(metaPath)
if err != nil {
continue
}

var container Container
if err := json.Unmarshal(data, &container); err != nil {
continue
}

if !allContainers && container.State != "running" {
continue
}

cmdStr := strings.Join(container.Command, " ")
if len(cmdStr) > 20 {
cmdStr = cmdStr[:20] + "..."
}

fmt.Printf("%-14s %-8s %-20s %-9s %-8s %-15s %s\n",
container.ID[:12],
container.Image,
cmdStr,
container.Created.Format("2006-01-02"),
container.State,
formatPorts(container.Ports),
container.Name)
}
}

func formatPorts(ports []PortMapping) string {
if len(ports) == 0 {
return ""
}
var parts []string
for _, p := range ports {
parts = append(parts, fmt.Sprintf("%s:%d->%d/%s", p.HostIP, p.HostPort, p.ContainerPort, p.Protocol))
}
return strings.Join(parts, ", ")
}

// handleImages lists images
func handleImages(args []string) {
fmt.Println("REPOSITORY   TAG   IMAGE ID   CREATED   SIZE")

home := getTerpodmanHome()
imagesPath := filepath.Join(home, imagesDir)

entries, err := os.ReadDir(imagesPath)
if err != nil {
return // No images yet
}

for _, entry := range entries {
if !strings.HasSuffix(entry.Name(), ".json") {
continue
}

metaPath := filepath.Join(imagesPath, entry.Name())
data, err := os.ReadFile(metaPath)
if err != nil {
continue
}

var image Image
if err := json.Unmarshal(data, &image); err != nil {
continue
}

sizeStr := formatSize(image.Size)
fmt.Printf("%-13s %-6s %-13s %-9s %s\n",
image.Name,
image.Tag,
image.ID[:12],
image.Created.Format("2006-01-02"),
sizeStr)
}
}

func formatSize(bytes int64) string {
const unit = 1024
if bytes < unit {
return fmt.Sprintf("%dB", bytes)
}
div, exp := int64(unit), 0
for n := bytes / unit; n >= unit; n /= unit {
div *= unit
exp++
}
return fmt.Sprintf("%.1f%cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// handleStop stops a container
func handleStop(args []string) {
if len(args) < 1 {
fmt.Fprintln(os.Stderr, "Error: container name or ID required")
os.Exit(1)
}

containerID := args[0]
fmt.Printf("Stopping container %s...\n", containerID)
fmt.Println("Note: Full stop implementation requires proot process management.")
}

// handleStart starts a container
func handleStart(args []string) {
if len(args) < 1 {
fmt.Fprintln(os.Stderr, "Error: container name or ID required")
os.Exit(1)
}

containerID := args[0]
fmt.Printf("Starting container %s...\n", containerID)
fmt.Println("Note: Full start implementation requires proot process management.")
}

// handleRm removes a container
func handleRm(args []string) {
if len(args) < 1 {
fmt.Fprintln(os.Stderr, "Error: container name or ID required")
os.Exit(1)
}

for _, containerID := range args {
fmt.Printf("Removing container %s...\n", containerID)
}
fmt.Println("Note: Full rm implementation requires metadata and rootfs cleanup.")
}

// handleRmi removes an image
func handleRmi(args []string) {
if len(args) < 1 {
fmt.Fprintln(os.Stderr, "Error: image name or ID required")
os.Exit(1)
}

for _, imageID := range args {
fmt.Printf("Removing image %s...\n", imageID)
}
fmt.Println("Note: Full rmi implementation requires image metadata and rootfs cleanup.")
}

// handleExec executes a command in a running container
func handleExec(args []string) {
if len(args) < 2 {
fmt.Fprintln(os.Stderr, "Error: container and command required")
fmt.Fprintln(os.Stderr, "Usage: terpodman exec <container> <command>")
os.Exit(1)
}

containerID := args[0]
command := args[1:]

fmt.Printf("Executing command in container %s: %s\n", containerID, strings.Join(command, " "))
fmt.Println("Note: In terpodman, exec runs a fresh proot against the same rootfs.")
fmt.Println("This differs from true container exec which shares the namespace.")
}

// handleLogs shows container logs
func handleLogs(args []string) {
if len(args) < 1 {
fmt.Fprintln(os.Stderr, "Error: container name or ID required")
os.Exit(1)
}

containerID := args[0]

home := getTerpodmanHome()
containersPath := filepath.Join(home, containersDir)

// Find container metadata
var logPath string
entries, _ := os.ReadDir(containersPath)
for _, entry := range entries {
if !strings.HasSuffix(entry.Name(), ".json") {
continue
}

metaPath := filepath.Join(containersPath, entry.Name())
data, err := os.ReadFile(metaPath)
if err != nil {
continue
}

var container Container
if err := json.Unmarshal(data, &container); err != nil {
continue
}

if container.ID == containerID || container.Name == containerID {
logPath = container.LogPath
break
}
}

if logPath == "" {
fmt.Fprintf(os.Stderr, "Error: container %s not found\n", containerID)
os.Exit(1)
}

file, err := os.Open(logPath)
if err != nil {
fmt.Fprintf(os.Stderr, "Error reading logs: %v\n", err)
os.Exit(1)
}
defer file.Close()

io.Copy(os.Stdout, file)
}

// handleBuild builds an image from Dockerfile
func handleBuild(args []string) {
var tagName string
dockerfilePath := "Dockerfile"
buildContext := "."

for i, arg := range args {
switch arg {
case "-t", "--tag":
if i+1 < len(args) {
tagName = args[i+1]
}
case "-f", "--file":
if i+1 < len(args) {
dockerfilePath = args[i+1]
}
}
if !strings.HasPrefix(arg, "-") {
buildContext = arg
}
}

fmt.Printf("Building image from %s...\n", buildContext)
if tagName != "" {
fmt.Printf("Tag: %s\n", tagName)
}
fmt.Printf("Dockerfile: %s\n", dockerfilePath)
fmt.Println("\nNote: Full build implementation requires:")
fmt.Println("- Dockerfile parsing (similar to podman/buildah)")
fmt.Println("- Layer creation in user-space")
fmt.Println("- proot-based build environment")
}

// handleVolume manages volumes
func handleVolume(args []string) {
if len(args) < 1 {
fmt.Fprintln(os.Stderr, "Usage: terpodman volume <create|ls|rm|inspect> [options]")
os.Exit(1)
}

subcmd := args[0]

switch subcmd {
case "create":
if len(args) < 2 {
fmt.Fprintln(os.Stderr, "Error: volume name required")
os.Exit(1)
}
volumeName := args[1]
home := getTerpodmanHome()
volumePath := filepath.Join(home, volumesDir, volumeName)
if err := os.MkdirAll(volumePath, 0755); err != nil {
fmt.Fprintf(os.Stderr, "Error creating volume: %v\n", err)
os.Exit(1)
}
fmt.Printf("Volume created: %s\n", volumeName)

case "ls", "list":
fmt.Println("DRIVER    VOLUME NAME")
home := getTerpodmanHome()
volumesPath := filepath.Join(home, volumesDir)
entries, _ := os.ReadDir(volumesPath)
for _, entry := range entries {
if entry.IsDir() {
fmt.Printf("local     %s\n", entry.Name())
}
}

case "rm", "remove":
if len(args) < 2 {
fmt.Fprintln(os.Stderr, "Error: volume name required")
os.Exit(1)
}
volumeName := args[1]
home := getTerpodmanHome()
volumePath := filepath.Join(home, volumesDir, volumeName)
if err := os.RemoveAll(volumePath); err != nil {
fmt.Fprintf(os.Stderr, "Error removing volume: %v\n", err)
os.Exit(1)
}
fmt.Printf("Volume removed: %s\n", volumeName)

case "inspect":
if len(args) < 2 {
fmt.Fprintln(os.Stderr, "Error: volume name required")
os.Exit(1)
}
volumeName := args[1]
home := getTerpodmanHome()
volumePath := filepath.Join(home, volumesDir, volumeName)

volume := Volume{
Name:       volumeName,
Driver:     "local",
Mountpoint: volumePath,
CreatedAt:  time.Now(),
}

data, _ := json.MarshalIndent(volume, "", "  ")
fmt.Println(string(data))

default:
fmt.Fprintf(os.Stderr, "Unknown volume subcommand: %s\n", subcmd)
os.Exit(1)
}
}

// handlePod manages pods
func handlePod(args []string) {
if len(args) < 1 {
fmt.Fprintln(os.Stderr, "Usage: terpodman pod <create|ls|rm|inspect> [options]")
os.Exit(1)
}

subcmd := args[0]

switch subcmd {
case "create":
var podName string
for i, arg := range args {
if arg == "--name" && i+1 < len(args) {
podName = args[i+1]
break
}
}
if podName == "" {
fmt.Fprintln(os.Stderr, "Error: --name required for pod create")
os.Exit(1)
}
fmt.Printf("Creating pod: %s\n", podName)
fmt.Println("Note: Pod implementation requires shared namespace simulation via proot.")

case "ls", "list":
fmt.Println("POD ID      POD NAME   STATUS   CONTAINERS")
home := getTerpodmanHome()
podsPath := filepath.Join(home, podsDir)
entries, _ := os.ReadDir(podsPath)
for _, entry := range entries {
if !strings.HasSuffix(entry.Name(), ".json") {
continue
}
metaPath := filepath.Join(podsPath, entry.Name())
data, _ := os.ReadFile(metaPath)
var pod Pod
if err := json.Unmarshal(data, &pod); err == nil {
fmt.Printf("%-12s %-11s %-9s %d\n", 
pod.ID[:12], pod.Name, pod.State, len(pod.Containers))
}
}

case "rm", "remove":
if len(args) < 2 {
fmt.Fprintln(os.Stderr, "Error: pod name or ID required")
os.Exit(1)
}
podID := args[1]
fmt.Printf("Removing pod: %s\n", podID)

case "inspect":
if len(args) < 2 {
fmt.Fprintln(os.Stderr, "Error: pod name or ID required")
os.Exit(1)
}
podID := args[1]
fmt.Printf("Inspecting pod: %s\n", podID)

default:
fmt.Fprintf(os.Stderr, "Unknown pod subcommand: %s\n", subcmd)
os.Exit(1)
}
}

// handleCompose manages compose projects
func handleCompose(args []string) {
if len(args) < 1 {
fmt.Fprintln(os.Stderr, "Usage: terpodman compose <up|down|ps> [options]")
os.Exit(1)
}

subcmd := args[0]

switch subcmd {
case "up":
fmt.Println("Starting compose project...")
fmt.Println("Note: Compose requires YAML parsing and multi-container orchestration.")

case "down":
fmt.Println("Stopping compose project...")

case "ps":
fmt.Println("COMPOSE PROJECT   SERVICE   CONTAINER ID   STATUS")

default:
fmt.Fprintf(os.Stderr, "Unknown compose subcommand: %s\n", subcmd)
os.Exit(1)
}
}

// handleInfo shows system information
func handleInfo(args []string) {
fmt.Println(`terpodman version 0.1.0-beta

What doesn't work (and never will, without root):
- --memory / --cpus — no cgroups on Android non-root. Flags are accepted and ignored.
- Real bridge networking — no network namespaces. Containers share the host network.
- Seccomp profiles — useless without PID namespaces.
- Overlayfs — no mount(). Each container gets a full copy of its image's rootfs.
- True exec into a running container — terpodman exec runs a fresh proot against the same rootfs.

Architecture:
- Uses proot for user-space syscall interception
- No kernel namespace support
- All state stored in ~/.local/share/terpodman/
- Port forwarding via user-space TCP proxy`)
}
