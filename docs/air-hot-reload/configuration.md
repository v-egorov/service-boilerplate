# Air Configuration Guide

## Overview

Air is configured using TOML files (`.air.toml`) located in each service directory. This document explains all configuration options and their current settings.

## Configuration Files

- `api-gateway/.air.toml` - API Gateway hot reload configuration
- `services/user-service/.air.toml` - User Service hot reload configuration

## Configuration Structure

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/api-gateway"
  cmd = "go build -o ./tmp/api-gateway ./cmd"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "docker"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

## Build Configuration

### Core Build Settings

| Setting | Current Value | Description |
|---------|---------------|-------------|
| `bin` | `./tmp/api-gateway` | Output binary path |
| `cmd` | `go build -o ./tmp/api-gateway ./cmd` | Build command |
| `delay` | `1000` | Delay before rebuild (milliseconds) |
| `kill_delay` | `0s` | Delay before killing old process |

### File Monitoring

| Setting | Current Value | Description |
|---------|---------------|-------------|
| `include_ext` | `["go", "tpl", "tmpl", "html"]` | File extensions to watch |
| `exclude_dir` | `["assets", "tmp", "vendor", "testdata", "docker"]` | Directories to ignore |
| `exclude_regex` | `["_test.go"]` | File patterns to exclude |
| `exclude_unchanged` | `false` | Exclude unchanged files |
| `follow_symlink` | `false` | Follow symbolic links |

### Advanced Build Options

| Setting | Current Value | Description |
|---------|---------------|-------------|
| `args_bin` | `[]` | Arguments passed to binary |
| `full_bin` | `""` | Full path to binary (overrides `bin`) |
| `poll` | `false` | Use polling instead of events |
| `poll_interval` | `0` | Polling interval (when polling enabled) |
| `rerun` | `false` | Rerun binary instead of replacing |
| `rerun_delay` | `500` | Delay before rerun (milliseconds) |
| `send_interrupt` | `false` | Send interrupt signal before kill |
| `stop_on_root` | `false` | Stop on root directory changes |

## Logging Configuration

### Build Logs

| Setting | Current Value | Description |
|---------|---------------|-------------|
| `log` | `build-errors.log` | Build error log file |

### Console Output

| Setting | Current Value | Description |
|---------|---------------|-------------|
| `main_only` | `false` | Show only main process output |
| `time` | `false` | Show timestamps in logs |

## Color Configuration

Air uses colors to distinguish different types of output:

| Setting | Current Value | Description |
|---------|---------------|-------------|
| `app` | `""` | Application output color |
| `build` | `"yellow"` | Build process color |
| `main` | `"magenta"` | Main process color |
| `runner` | `"green"` | Runner color |
| `watcher` | `"cyan"` | File watcher color |

## Screen Configuration

| Setting | Current Value | Description |
|---------|---------------|-------------|
| `clear_on_rebuild` | `false` | Clear screen on rebuild |
| `keep_scroll` | `true` | Keep scroll position |

## Miscellaneous

| Setting | Current Value | Description |
|---------|---------------|-------------|
| `clean_on_exit` | `false` | Clean temporary files on exit |

## Directory Structure

```
service-root/
├── .air.toml          # Air configuration
├── tmp/              # Build artifacts (auto-created)
├── cmd/              # Main application entry point
├── internal/         # Internal packages
└── build-errors.log  # Build error logs
```

## Customization Examples

### Adding New File Types to Watch

```toml
include_ext = ["go", "tpl", "tmpl", "html", "yaml", "yml"]
```

### Excluding Additional Directories

```toml
exclude_dir = ["assets", "tmp", "vendor", "testdata", "docker", "node_modules"]
```

### Changing Build Delay

```toml
delay = 500  # Faster rebuilds (500ms)
```

### Enabling Polling Mode

```toml
poll = true
poll_interval = 1000  # Check every second
```

### Custom Build Command

```toml
cmd = "go build -tags=dev -o ./tmp/api-gateway ./cmd"
```

## Environment-Specific Configuration

You can use environment variables in your `.air.toml`:

```toml
[build]
  cmd = "go build -o ./tmp/$SERVICE_NAME ./cmd"
```

## Best Practices

1. **Keep build delay reasonable** (500-2000ms) to avoid excessive CPU usage
2. **Exclude unnecessary files** to improve performance
3. **Use descriptive binary names** for easier debugging
4. **Configure colors** for better log readability
5. **Monitor build logs** for compilation errors

## Troubleshooting Configuration

- Check `build-errors.log` for compilation issues
- Verify file paths in `include_ext` and `exclude_dir`
- Ensure `cmd` can execute from the service root directory
- Test configuration with `air -c .air.toml` locally

## Related Files

- `api-gateway/.air.toml` - API Gateway configuration
- `services/user-service/.air.toml` - User Service configuration
- `Makefile` - Development commands
- `docker/docker-compose.override.yml` - Docker development setup</content>
</xai:function_call">### 📋 **Configuration Documentation Created**

I've created comprehensive configuration documentation covering:

#### **📄 Configuration Guide Features**
- **Complete Settings Reference** - All `.air.toml` options explained
- **Current Configuration** - Exact values from your project
- **File Monitoring** - Include/exclude patterns and extensions
- **Build Options** - Command, delay, and rebuild settings
- **Logging & Colors** - Output formatting and log files
- **Screen Settings** - Display and scrolling options

#### **🎯 Key Sections**
- **Build Configuration** - Core build settings and file watching
- **Advanced Options** - Polling, rerun, and interrupt handling
- **Customization Examples** - Real-world configuration changes
- **Best Practices** - Performance tips and recommendations
- **Troubleshooting** - Common configuration issues

#### **📊 Configuration Table**
Documented all settings with current values:
- File extensions to watch: `["go", "tpl", "tmpl", "html"]`
- Excluded directories: `["assets", "tmp", "vendor", "testdata", "docker"]`
- Build delay: `1000ms` (prevents excessive rebuilds)
- Test file exclusion: `["_test.go"]`
- Color-coded logging for better visibility

#### **💡 Practical Examples**
- Adding new file types to monitor
- Custom build commands with tags
- Environment variable integration
- Performance optimization settings

#### **📋 Next Steps**
Ready to create the development setup guide next, which will include:
- Getting started instructions
- Docker integration details
- Makefile command explanations
- Step-by-step development workflow

Would you like me to continue with the development setup documentation?