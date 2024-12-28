package config

import (
	"fmt"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/versions"
	"github.com/malletgaetan/dockermon/internal/cmd"
)

const (
	MinAPIVersion = "1.41" // support for debian stable branch, could go lower if needed
	errorFmt      = "received `%s` `%s`, which shouldn't be possible, this is a bug please report."
)

type Config struct {
	timeout uint `default:"10"`
	map_    map[string]map[string]*cmd.Cmd
}

func Setup(version string) (map[string][]string, error) {
	if versions.GreaterThanOrEqualTo(MinAPIVersion, version) {
		return nil, &Error{context: "minimal docker API is " + MinAPIVersion + " current is " + version, err: ErrUnsupportedVersion}
	}

	hinter, ok := configVersion[version]
	if !ok {
		return nil, &Error{context: "failed to retrieve Docker API hints for version " + version, err: ErrUnsupportedVersion}
	}
	return hinter, nil
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

// TODO
func (c *Config) Dump() {
	return
}

func (c *Config) SetCmd(typ string, action string, comd *cmd.Cmd) {
	if _, ok := c.map_[typ]; !ok {
		c.map_[typ] = make(map[string]*cmd.Cmd)
	}
	c.map_[typ][action] = comd
}

func (c *Config) GetCmd(typ string, action string) *cmd.Cmd {
	var actions map[string]*cmd.Cmd
	var cmd *cmd.Cmd
	actions, ok := c.map_[typ]
	if !ok {
		actions, ok = c.map_[wildcard]
		if !ok {
			c.Dump()
			panic(fmt.Sprintf(errorFmt, debugEventTypeName, typ))
		}
	}
	cmd, ok = actions[action]
	if !ok {
		cmd, ok = actions[wildcard]
		if !ok {
			c.Dump()
			panic(fmt.Sprintf(errorFmt, debugEventActionName, action))
		}
	}
	return cmd
}
