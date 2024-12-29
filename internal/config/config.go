package config

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

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

func (c *Config) Dump() {
	fmt.Println("Internal configuration dump:")
	for eventType, actionMap := range c.map_ {
		actions := keys(actionMap)
		sort.Strings(actions)
		for _, action := range actions {
			cmd := actionMap[action]
			timeout := ""
			var t uint = 0
			if cmd.Timeout != 0 {
				t = cmd.Timeout
			} else if c.timeout != 0 {
				t = c.timeout
			}
			if t != 0 {
				timeout += strconv.FormatUint(uint64(t), 10)
			}
			fmt.Printf("%s%s%s%s%s%s%s\n", eventType, delimiter, action, delimiter, timeout, delimiter, strings.Join(cmd.Args, ","))
		}
	}
}

func (c *Config) SetCmd(typ string, action string, comd *cmd.Cmd) {
	if _, ok := c.map_[typ]; !ok {
		c.map_[typ] = make(map[string]*cmd.Cmd)
	}
	c.map_[typ][action] = comd
}

func (c *Config) GetCmd(typ string, action string) (*cmd.Cmd, error) {
	var actions map[string]*cmd.Cmd
	var cmd *cmd.Cmd
	actions, ok := c.map_[typ]
	if !ok {
		actions, ok = c.map_[wildcard]
		if !ok {
			return nil, &Error{context: fmt.Sprintf(errorFmt, debugEventTypeName, typ), err: ErrUnimplemented}
		}
	}
	cmd, ok = actions[action]
	if !ok {
		cmd, ok = actions[wildcard]
		if !ok {
			return nil, &Error{context: fmt.Sprintf(errorFmt, debugEventActionName, typ), err: ErrUnimplemented}
		}
	}
	return cmd, nil
}
