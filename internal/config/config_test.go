package config

import (
	"testing"

	"github.com/malletgaetan/dockermon/internal/cmd"
	"github.com/stretchr/testify/assert"
)

func TestSetCmd(t *testing.T) {
	tests := []struct {
		name     string
		typ      string
		action   string
		cmd      *cmd.Cmd
		validate func(*testing.T, *Config)
	}{
		{
			name:   "set new type and action",
			typ:    "container",
			action: "start",
			cmd: &cmd.Cmd{
				Args:    []string{"/usr/bin/notify"},
				Timeout: 5,
			},
			validate: func(t *testing.T, c *Config) {
				assert.Len(t, c.map_, 1)
				assert.Len(t, c.map_["container"], 1)
				assert.Equal(t, uint(5), c.map_["container"]["start"].Timeout)
				assert.Equal(t, []string{"/usr/bin/notify"}, c.map_["container"]["start"].Args)
			},
		},
		{
			name:   "override existing command",
			typ:    "container",
			action: "start",
			cmd: &cmd.Cmd{
				Args:    []string{"/usr/bin/new-notify"},
				Timeout: 10,
			},
			validate: func(t *testing.T, c *Config) {
				assert.Len(t, c.map_, 1)
				assert.Len(t, c.map_["container"], 1)
				assert.Equal(t, uint(10), c.map_["container"]["start"].Timeout)
				assert.Equal(t, []string{"/usr/bin/new-notify"}, c.map_["container"]["start"].Args)
			},
		},
		{
			name:   "add new action to existing type",
			typ:    "container",
			action: "stop",
			cmd: &cmd.Cmd{
				Args:    []string{"/usr/bin/stop-notify"},
				Timeout: 3,
			},
			validate: func(t *testing.T, c *Config) {
				assert.Len(t, c.map_, 1)
				assert.Len(t, c.map_["container"], 2)
				assert.Equal(t, uint(3), c.map_["container"]["stop"].Timeout)
				assert.Equal(t, []string{"/usr/bin/stop-notify"}, c.map_["container"]["stop"].Args)
			},
		},
		{
			name:   "set wildcard action",
			typ:    "network",
			action: "*",
			cmd: &cmd.Cmd{
				Args:    []string{"/usr/bin/network-monitor"},
				Timeout: 0,
			},
			validate: func(t *testing.T, c *Config) {
				assert.Contains(t, c.map_, "network")
				assert.Contains(t, c.map_["network"], "*")
				assert.Equal(t, uint(0), c.map_["network"]["*"].Timeout)
				assert.Equal(t, []string{"/usr/bin/network-monitor"}, c.map_["network"]["*"].Args)
			},
		},
	}

	config := &Config{
		map_: make(map[string]map[string]*cmd.Cmd),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.SetCmd(tt.typ, tt.action, tt.cmd)
			tt.validate(t, config)
		})
	}
}

func TestGetCmd(t *testing.T) {
	config := &Config{
		map_: make(map[string]map[string]*cmd.Cmd),
	}

	specificCmd := &cmd.Cmd{Args: []string{"/usr/bin/specific"}, Timeout: 5}
	wildcardActionCmd := &cmd.Cmd{Args: []string{"/usr/bin/wildcard-action"}, Timeout: 3}

	config.SetCmd("container", "start", specificCmd)
	config.SetCmd("container", "*", wildcardActionCmd)

	tests := []struct {
		name        string
		typ         string
		action      string
		expectedCmd *cmd.Cmd
		expectError bool
		errorType   error
	}{
		{
			name:        "get specific command",
			typ:         "container",
			action:      "start",
			expectedCmd: specificCmd,
			expectError: false,
		},
		{
			name:        "fallback to wildcard action",
			typ:         "container",
			action:      "stop",
			expectedCmd: wildcardActionCmd,
			expectError: false,
		},
		{
			name:        "unknown type",
			typ:         "unknown",
			action:      "start",
			expectedCmd: nil,
			expectError: true,
			errorType:   ErrUnimplemented,
		},
		{
			name:        "unknown action without wildcard",
			typ:         "network",
			action:      "connect",
			expectedCmd: nil,
			expectError: true,
			errorType:   ErrUnimplemented,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := config.GetCmd(tt.typ, tt.action)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
				assert.Nil(t, cmd)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCmd, cmd)
			}
		})
	}
}
