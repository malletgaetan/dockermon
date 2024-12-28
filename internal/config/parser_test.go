package config

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/malletgaetan/dockermon/internal/cmd"
	"github.com/stretchr/testify/assert"
)

func FuzzParseConfig(f *testing.F) {
	corpusBytes, err := os.ReadFile("../../configs/corpus.conf")
	if err != nil {
		f.Fatal(err)
	}

	f.Add(corpusBytes)

	f.Fuzz(func(t *testing.T, data []byte) {
		hints, _ := configVersion["1.47"]
		scanner := bufio.NewScanner(bytes.NewReader(data))
		config, err := ParseConfig(scanner, hints)

		if err != nil {
			if config != nil {
				t.Errorf("got non-nil config with error: %v", err)
			}
			return
		}

		if config == nil {
			t.Error("got nil config without error")
		}
	})
}

func TestParseLine(t *testing.T) {
	hints := map[string][]string{
		"container": {"start", "die"},
		"network":   {"connect", "disconnect"},
	}

	tests := []struct {
		name        string
		input       string
		expectedErr string
		validate    func(*testing.T, *Config)
	}{
		{
			name:  "empty line",
			input: "",
			validate: func(t *testing.T, c *Config) {
				assert.Empty(t, c.map_)
			},
		},
		{
			name:  "comment line",
			input: "# This is a comment",
			validate: func(t *testing.T, c *Config) {
				assert.Empty(t, c.map_)
			},
		},
		{
			name:  "valid container start handler",
			input: "container::start::5::'/usr/bin/slack_notify','info'",
			validate: func(t *testing.T, c *Config) {
				cmd := c.map_["container"]["start"]
				assert.NotNil(t, cmd)
				assert.Equal(t, uint(5), cmd.Timeout)
				assert.Equal(t, []string{"/usr/bin/slack_notify", "info"}, cmd.Args)
			},
		},
		{
			name:  "valid wildcard action handler",
			input: "container::*::5::'/usr/bin/log_event'",
			validate: func(t *testing.T, c *Config) {
				cmd := c.map_["container"]["*"]
				assert.NotNil(t, cmd)
				assert.Equal(t, uint(5), cmd.Timeout)
				assert.Equal(t, []string{"/usr/bin/log_event"}, cmd.Args)
			},
		},
		{
			name:  "valid handler with empty timeout",
			input: "network::*::::'/usr/bin/network_monitor'",
			validate: func(t *testing.T, c *Config) {
				cmd := c.map_["network"]["*"]
				assert.NotNil(t, cmd)
				assert.Equal(t, uint(0), cmd.Timeout)
				assert.Equal(t, []string{"/usr/bin/network_monitor"}, cmd.Args)
			},
		},
		{
			name:  "valid handler with escaped quotes",
			input: "container::die::5::'/usr/bin/alert','\\'error\\''",
			validate: func(t *testing.T, c *Config) {
				cmd := c.map_["container"]["die"]
				assert.NotNil(t, cmd)
				assert.Equal(t, []string{"/usr/bin/alert", "'error'"}, cmd.Args)
			},
		},
		{
			name:        "invalid type",
			input:       "invalid_type::start::5::'/usr/bin/cmd'",
			expectedErr: "invalid type `invalid_type`",
		},
		{
			name:        "invalid action",
			input:       "container::invalid_action::5::'/usr/bin/cmd'",
			expectedErr: "invalid action `invalid_action`",
		},
		{
			name:        "wildcard type",
			input:       "*::start::5::'/usr/bin/cmd'",
			expectedErr: "type can't be wildcard",
		},
		{
			name:        "missing action delimiter",
			input:       "container:start::5::'/usr/bin/cmd'",
			expectedErr: "invalid type `container:start`",
		},
		{
			name:        "missing timeout delimiter",
			input:       "container::start:5::'/usr/bin/cmd'",
			expectedErr: "invalid action `start:5`",
		},
		{
			name:        "invalid timeout value",
			input:       "container::start::abc::'/usr/bin/cmd'",
			expectedErr: "strconv.ParseUint",
		},
		{
			name:        "missing command argument start quote",
			input:       "container::start::5::/usr/bin/cmd'",
			expectedErr: "no start delimiter found for command argument",
		},
		{
			name:        "missing command argument end quote",
			input:       "container::start::5::'/usr/bin/cmd",
			expectedErr: "no end delimiter found for command argument",
		},
		{
			name:        "missing delimiter event action",
			input:       "container::start:5:'/usr/bin/cmd';'arg'",
			expectedErr: "no delimiter found after event action",
		},
		{
			name:        "missing delimiter event type",
			input:       "container:start:5:'/usr/bin/cmd';'arg'",
			expectedErr: "no delimiter found after event type",
		},
		{
			name:        "invalid argument delimiter",
			input:       "container::start::5::'/usr/bin/cmd';'arg'",
			expectedErr: "expected delimiter after argument",
		},
		{
			name:  "valid global timeout setting",
			input: "timeout=30",
			validate: func(t *testing.T, c *Config) {
				assert.Equal(t, uint(30), c.timeout)
			},
		},
		{
			name:        "invalid global setting",
			input:       "invalid=value",
			expectedErr: "unknown global setting",
		},
		{
			name:        "invalid global timeout value",
			input:       "timeout=abc",
			expectedErr: "strconv.ParseUint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := bufio.NewScanner(strings.NewReader(tt.input))
			config := &Config{map_: make(map[string]map[string]*cmd.Cmd)}

			parser := &parser{
				scanner: scanner,
				config:  config,
				hints:   hints,
				pos:     Position{row: 1},
			}

			scanner.Scan()
			err := parser.parseLine()

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, config)
				}
			}
		})
	}
}
