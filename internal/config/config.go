package config

import (
	"github.com/docker/docker/api/types/filters"
	"github.com/malletgaetan/dockermon/internal/cmd"
)

const (
	MinAPIVersion = "1.41" // support for debian stable branch, could go lower if needed
)

type Config struct {
	timeout uint
	version string
	map_    map[string]map[string]*cmd.Cmd
}

func (c *Config) Filters() filters.Args {
	filterArgs := filters.NewArgs()

	for typ, innerMap := range c.map_ {
		filterArgs.Add("type", typ)
		if _, ok := innerMap[wildcard]; ok {
			continue
		}

		for action, _ := range innerMap {
			filterArgs.Add("event", action)
		}
	}
	return filterArgs
}

func (c *Config) Dump() {
	return
}

func (c *Config) SetCmd(typ string, action string, comd *cmd.Cmd) {
	if _, ok := c.map_[typ]; !ok {
		c.map_[typ] = make(map[string]*cmd.Cmd)
	}
	c.map_[typ][action] = comd
}

func (c *Config) GetCmd(typ string, action string) (*cmd.Cmd, error) {
	configError := &ConfigError{
		err: ErrUnimplemented,
	}
	var actions map[string]*cmd.Cmd
	var cmd *cmd.Cmd
	actions, ok := c.map_[typ]
	if !ok {
		actions, ok = c.map_[wildcard]
		if !ok {
			c.Dump()
			configError.message = "typ '" + typ + "' not set, this is a bug please report."
			return nil, configError
		}
	}
	cmd, ok = actions[action]
	if !ok {
		cmd, ok = actions[wildcard]
		if !ok {
			c.Dump()
			configError.message = "action '" + action + "' not set, this is a bug please report."
			return nil, configError
		}
	}
	return cmd, nil
}
