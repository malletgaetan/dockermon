# Dockermon

A lightweight, flexible tool for attaching custom hooks to Docker events. Monitor your containers and trigger actions based on container lifecycle events with minimal setup.

## Features

- Monitor Docker container lifecycle events
- Execute custom commands on specific events
- Configurable timeouts per command or globally
- Support for wildcard event matching
- Simple configuration file format

## Usage

Start Dockermon by providing your configuration file:

```bash
dockermon -c <config_filepath>
```

## Configuration

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

### Example Configuration

```bash
# Send Slack notification on container start (5s timeout)
container::start::5::'/usr/bin/slack_notify','info'

# Execute command for all container events
container::*::5::'/usr/bin/slack_notify','info'

# Special handler for container die events
container::die::5::'/usr/bin/slack_notify','error'

# Network events handler with no timeout
network::*::::'/usr/bin/stuff'
```

## Event Types

- `container` - Container lifecycle events (start, die, etc.)
- `network` - Network-related events
