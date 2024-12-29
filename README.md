# Dockermon

A lightweight, flexible tool for attaching custom hooks to Docker events. Monitor your containers and trigger actions based on container lifecycle events with minimal setup.

## Features

- Real-time monitoring of Docker container lifecycle events
- Execute custom commands on specific events with event data passed through stdin
- Robust error handling for misconfigured event types or actions
- Configurable timeouts (both global and per-command)
- Support for wildcard event matching
- Simple, human-readable configuration file format
- Validation against Docker API version compatibility

## Quick Start

### Installation

Build from source:
```bash
make
```

Or install using pre-built binaries:
```bash
curl -fsSL https://raw.githubusercontent.com/malletgaetan/dockermon/main/install.sh | sudo bash
```

### Basic Usage

Start Dockermon by providing your configuration file:
```bash
dockermon -c <config_filepath>
```

## Configuration Guide

### File Format

Dockermon uses a simple text configuration file format:

```bash
# Global timeout in seconds (optional, default: no timeout)
timeout=60

# Format: type::action::timeout::command
# type    - Event type (container, network, etc.)
# action  - Event action or * for wildcard
# timeout - Command timeout in seconds (optional)
# command - Command to execute with arguments
```

### Event Data

Commands receive event data through stdin in JSON format. Example event data:

```json
{
  "Type": "container",
  "Action": "start",
  "Actor": {
    "ID": "abc123...",
    "Attributes": {
      "name": "my-container",
      "image": "nginx:latest"
    }
  },
  "time": 1234567890
}
```

### Example Configurations

```bash
# Send Slack notification on container start (5s timeout)
container::start::5::'/usr/bin/slack_notify','info'

# Execute command for all container events
container::*::5::'/usr/bin/log_event'

# Special handler for container die events
container::die::5::'/usr/bin/alert','error'

# Network events handler with no timeout
network::*::::'/usr/bin/network_monitor'
```

## Error Handling

Dockermon provides clear error messages for invalid configurations. For example:

```bash
dockermon -c configs/corpus.conf
```

Will yield:
```
Error: failed to parse configuration
Parsing Error: invalid action `starts` on type `container`, use one of: [attach commit copy create destroy detach die exec_create exec_detach exec_start exec_die export health_status kill oom pause rename resize restart start stop top unpause update prune]
    6 |container::starts::5::'/usr/bin/notify'
      |           ^^^^^^                      
Parsing Error: invalid action `die` on type `network`, use one of: [create connect disconnect destroy update remove prune]
    8 |network::die::0::'/usr/bin/log'
      |         ^^^                   
Parsing Error: invalid type `storage`, use one of: [network service node secret config container image volume]
    9 |storage::*::::'/usr/bin/storage'
      |^^^^^^^
```

## Supported Events

Dockermon supports all event types and actions from the Docker Engine API. For a complete list, refer to the [Docker API documentation](https://docs.docker.com/reference/api/engine/version/v1.47/#tag/System/operation/SystemEvents).
