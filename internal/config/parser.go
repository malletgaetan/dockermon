package config

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/versions"
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

// TODO: fix
func keys[T any](m map[string]T) []string {
	arr := make([]string, len(m))
	i := 0
	for k, _ := range m {
		arr[i] = k
		i += 1
	}
	return arr
}

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

func parseHandlerLine(config *Config, hinter map[string][]string, line string) error {
	// parse event type
	typ_end := strings.Index(line, delimiter)
	if typ_end == -1 {
		return &ConfigError{message: "no delimiter found after event type", err: ErrMalformed}
	}
	typ := line[:typ_end]
	if typ == wildcard {
		return &ConfigError{message: "type can't be wildcare, use wildcare only for actions", err: ErrUnimplemented}
	}
	possibleActions, ok := hinter[typ]
	if !ok {
		return &ConfigError{message: fmt.Sprintf("event of type `%v` does not exist on your current docker version, use one of: %v", typ, keys(hinter)), err: ErrUnimplemented}
	}

	// parse event action
	action_end := strings.Index(line[typ_end+2:], delimiter)
	if action_end == -1 {
		return &ConfigError{message: "no delimiter found after event action", err: ErrMalformed}
	}
	action_end = action_end + typ_end + 2
	action := line[typ_end+2 : action_end]
	if action != wildcard && !slices.Contains(possibleActions, action) {
		return &ConfigError{message: fmt.Sprintf("action `%v` on type `%v` does not exist on your current docker version, use one of: %v", action, typ, possibleActions), err: ErrUnimplemented}
	}

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
		if j >= len(line) || line[j] != arg_container {
			return &ConfigError{message: "no start delimiter found for arg", err: ErrMalformed}
		}
		j++
		arg := ""
		for {
			if j >= len(line) {
				return &ConfigError{message: "no end start delimiter found in arg: `" + arg + "`", err: ErrMalformed}
			}
			if line[j] == '\\' && j+1 < len(line) {
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
			return &ConfigError{message: "expected delimiter after arg: `" + arg + "`", err: ErrMalformed}
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

func parseLine(config *Config, hinter map[string][]string, line string) error {
	if len(line) == 0 || line[0] == comment {
		return nil
	}

	i := strings.Index(line, global_delimiter)
	if i != -1 {
		return parseGlobalSettingLine(config, line, i)
	}
	return parseHandlerLine(config, hinter, line)
}

func ParseConfig(scanner *bufio.Scanner, version string) (*Config, error) {
	if versions.GreaterThanOrEqualTo(MinAPIVersion, version) {
		return nil, &ConfigError{message: "minimal docker API version supported is " + MinAPIVersion + " current is " + version}
	}

	hinter, ok := configVersion[version]
	if !ok {
		return nil, &ConfigError{message: "failed to found commands hinter for Docker API version " + version + ", this is a bug, please report."}
	}

	config := &Config{
		map_:    make(map[string]map[string]*cmd.Cmd),
		version: version,
	}

	errHappened := false
	for scanner.Scan() {
		err := parseLine(config, hinter, scanner.Text())
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
		return nil, ErrMalformed
	}
	return config, nil
}

func ParseConfigFile(filepath string, version string) (*Config, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	return ParseConfig(scanner, version)
}
