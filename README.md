# Dockermon

A lightweight, flexible tool for attaching custom hooks to Docker events. Monitor your containers and trigger actions based on container lifecycle events with minimal setup.

## Features

- Runtime errors on bad configured event type or action depending on your docker version
- Monitor Docker container lifecycle events
- Execute custom commands on specific events
- Event data through stdin
- Configurable timeouts per command or globally
- Support for wildcard event matching
- Simple configuration file format

## Usage

Build with:

```bash
make
```

Or download pre-built binaries with:

```bash
// TODO
```

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

Unlike docker events cli, dockermon handle invalid event types or actions:

```bash
gm@tower:~/code/dockermon$ ./dockermon -c configs/corpus.conf 
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

## Supported Events

See [Docker API](https://docs.docker.com/reference/api/engine/version/v1.47/#tag/System/operation/SystemEvents)