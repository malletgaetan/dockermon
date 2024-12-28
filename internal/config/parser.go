package config

import (
	"bufio"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/malletgaetan/dockermon/internal/cmd"
)

const (
	debugEventTypeName   = "Event Type"
	debugEventActionName = "Event Action"

	wildcard          = "*"
	delimiter         = "::"
	globalDelimiter   = "="
	timeoutIdentifier = "timeout"

	comment      = '#'
	argContainer = '\''
	argDelimiter = ','
)

type Position struct {
	row int
	col int
}

type parser struct {
	scanner *bufio.Scanner
	hints   map[string][]string
	config  *Config
	pos     Position
}

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

func (p *parser) parseGlobalSettingLine(line string, mid int) error {
	item := line[:mid]
	value := line[mid+1:]

	if item != timeoutIdentifier {
		return &Error{
			context: "unknown global setting: '" + item + "'",
			err:     ErrBadValue,
		}
	}

	timeoutNb, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return &Error{
			context: "invalid timeout value: '" + value + "'",
			err:     err,
		}
	}

	p.config.timeout = uint(timeoutNb)
	return nil
}

func (p *parser) parseHandlerLine(line string) error {
	// parse event type
	typ_end := strings.Index(line, delimiter)
	if typ_end == -1 {
		return &ParserError{line: line, context: "no delimiter found after event type", err: ErrBadSyntax, pos: p.pos, length: len(line)}
	}
	typ := line[:typ_end]
	if typ == wildcard {
		return &ParserError{line: line, context: "type can't be wildcard, use wildcard only for actions", err: ErrBadValue, pos: p.pos, length: len(typ)}
	}
	possibleActions, ok := p.hints[typ]
	if !ok {
		return &ParserError{line: line, context: fmt.Sprintf("invalid type `%v`, use one of: %v", typ, keys(p.hints)), err: ErrBadValue, pos: p.pos, length: len(typ)}
	}

	p.pos.col = typ_end + len(delimiter)

	// parse event action
	action_end := strings.Index(line[typ_end+len(delimiter):], delimiter)
	if action_end == -1 {
		return &ParserError{line: line, context: "no delimiter found after event action", err: ErrBadSyntax, pos: p.pos, length: len(line) - p.pos.col}
	}
	action_end = action_end + typ_end + len(delimiter)
	action := line[typ_end+2 : action_end]
	if action != wildcard && !slices.Contains(possibleActions, action) {
		return &ParserError{line: line, context: fmt.Sprintf("invalid action `%v` on type `%v`, use one of: %v", action, typ, possibleActions), err: ErrBadValue, pos: p.pos, length: len(action)}
	}

	p.pos.col = action_end + len(delimiter)

	// parse cmd timeout
	timeout_end := strings.Index(line[action_end+2:], delimiter)
	if timeout_end == -1 {
		return &ParserError{line: line, context: "no delimiter found after timeout", err: ErrBadSyntax, pos: p.pos, length: len(line) - p.pos.col}
	}
	var timeoutNb uint64
	timeout_end = timeout_end + action_end + len(delimiter)
	timeoutNb = 0
	timeout := line[action_end+2 : timeout_end]
	if timeout != "" {
		var err error
		timeoutNb, err = strconv.ParseUint(timeout, 10, 16)
		if err != nil {
			return &ParserError{line: line, context: err.Error(), err: err, pos: p.pos, length: len(timeout)}
		}
	}

	// parse cmd
	j := timeout_end + len(delimiter)
	args := []string{}
	for {
		p.pos.col = j
		if j >= len(line) || line[j] != argContainer {
			return &ParserError{line: line, context: "no start delimiter found for command argument", err: ErrBadSyntax, pos: p.pos, length: len(line) - p.pos.col}
		}
		j++
		arg := ""
		for {
			if j >= len(line) {
				return &ParserError{line: line, context: "no end delimiter found for command argument", err: ErrBadSyntax, pos: p.pos, length: len(line) - p.pos.col}
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
			p.pos.col = j
		}
		j++
		args = append(args, arg)
		if j == len(line) {
			break
		}
		p.pos.col = j
		if line[j] != argDelimiter {
			return &ParserError{line: line, context: "expected delimiter after argument", err: ErrBadSyntax, pos: p.pos, length: 1}
		}
		j++
	}

	p.config.SetCmd(typ, action, &cmd.Cmd{
		Name:    args[0],
		Args:    args[1:],
		Timeout: uint(timeoutNb),
	})

	return nil
}

func (p *parser) parseLine() error {
	line := p.scanner.Text()
	if len(line) == 0 || line[0] == comment {
		return nil
	}

	i := strings.Index(line, globalDelimiter)
	if i != -1 {
		return p.parseGlobalSettingLine(line, i)
	}
	return p.parseHandlerLine(line)
}

func ParseConfig(scanner *bufio.Scanner, hints map[string][]string) (*Config, error) {
	config := &Config{
		map_: make(map[string]map[string]*cmd.Cmd),
	}

	parser := &parser{
		pos:     Position{row: 0, col: 0},
		config:  config,
		scanner: scanner,
		hints:   hints,
	}

	var err error = nil
	for parser.scanner.Scan() {
		err = errors.Join(err, parser.parseLine())
		parser.pos.row += 1
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Join(err, &Error{context: "failed to read file buffer", err: err})
	}

	if err != nil {
		return nil, err
	}

	return config, nil
}
