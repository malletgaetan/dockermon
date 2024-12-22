package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/malletgaetan/dockermon/internal/cmd"
)

const (
	wildcard           = "*"
	delimiter          = "::"
	global_delimiter   = "="
	timeout_identifier = "timeout"

	comment       = '#'
	arg_container = '\''
	arg_delimiter = ','
)

func parseGlobalSettingLine(config *Config, line string, mid int) error {
	item := line[:mid]
	value := line[mid+1:]

	if item != timeout_identifier {
		return &ConfigError{
			message: "unknown global setting: '" + item + "'",
			err:     ErrMalformed,
		}
	}

	timeoutNb, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return &ConfigError{
			message: "invalid timeout value: '" + value + "'",
			err:     err,
		}
	}

	config.timeout = uint(timeoutNb)
	return nil
}

func parseHandlerLine(config *Config, line string) error {
	// parse event type
	typ_end := strings.Index(line, delimiter)
	if typ_end == -1 {
		return &ConfigError{message: "no delimiter found after event type", err: ErrMalformed}
	}
	typ := line[:typ_end]
	if typ == wildcard {
		return &ConfigError{message: "type can't be wildcare, use wildcare only for actions"}
	}

	// parse event action
	action_end := strings.Index(line[typ_end+2:], delimiter)
	if action_end == -1 {
		return &ConfigError{message: "no delimiter found after event action", err: ErrMalformed}
	}
	action_end = action_end + typ_end + 2
	action := line[typ_end+2 : action_end]

	// parse cmd timeout
	timeout_end := strings.Index(line[action_end+2:], delimiter)
	if timeout_end == -1 {
		return &ConfigError{message: "no delimiter found after timeout", err: ErrMalformed}
	}
	var timeoutNb uint64
	timeout_end = timeout_end + action_end + 2
	timeoutNb = 0
	timeout := line[action_end+2 : timeout_end]
	if timeout != "" {
		var err error
		timeoutNb, err = strconv.ParseUint(timeout, 10, 16)
		if err != nil {
			return err
		}
	}

	// parse cmd
	j := timeout_end + 2
	args := []string{}
	for {
		if line[j] != arg_container {
			return &ConfigError{message: "no start delimiter found for arg", err: ErrMalformed}
		}
		j++
		arg := ""
		for {
			if j >= len(line) {
				return &ConfigError{message: "no end start delimiter found in arg: '" + arg + "'", err: ErrMalformed}
			}
			if line[j] == '\\' {
				arg += string(line[j+1])
				j += 2
				continue
			}
			if line[j] == '\'' {
				break
			}
			arg += string(line[j])
			j++
		}
		j++
		args = append(args, arg)
		if j == len(line) {
			break
		}
		if line[j] != arg_delimiter {
			return &ConfigError{message: "expected delimiter after arg: '" + arg + "'", err: ErrMalformed}
		}
		j++
	}

	config.SetCmd(typ, action, &cmd.Cmd{
		Name:    args[0],
		Args:    args[1:],
		Timeout: uint(timeoutNb),
	})

	return nil
}

func parseLine(config *Config, line string) error {
	if len(line) == 0 || line[0] == comment {
		return nil
	}

	i := strings.Index(line, global_delimiter)
	if i != -1 {
		return parseGlobalSettingLine(config, line, i)
	}
	return parseHandlerLine(config, line)
}

func ParseConfig(filepath string) (*Config, error) {
	config := &Config{
		map_: make(map[string]map[string]*cmd.Cmd),
	}
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	errHappened := false
	for scanner.Scan() {
		err := parseLine(config, scanner.Text())
		if err != nil {
			fmt.Println("Error found in line [", scanner.Text(), "]:")
			fmt.Println("---- ", err.Error())
			errHappened = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if errHappened {
		return config, ErrMalformed
	}
	return config, nil
}
