# Dockermon

Dockermon is a lightweight, flexible tool that lets you easily attach custom hooks to Docker events. Monitor your containers and trigger actions based on container lifecycle events with minimal setup.


## Getting Started

Install
```bash
TODO
```

Start
```bash
dockermon -c <config_filepath>
```

## Configuration format

```conf
# this is a comment
# global timeout, if not set per command, this timeout will be used, default non timeout
timeout=60

# type::action::timeout::command
container::start::5::'/usr/bin/slack_notify','info'
# action can be a wildcard to match every possible actions
container::*::5::'/usr/bin/slack_notify','info'
# create a handler that will be executed specificaly for this type and action, wildcard will not be invoked
container::die::5::'/usr/bin/slack_notify','error'
# timeout can be unset
network::*::::'/usr/bin/stuff'
```
