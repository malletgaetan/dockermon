package config

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrUnsupportedVersion = errors.New("unsupported docker version")
	ErrUnimplemented      = errors.New("unimplemented command type")
)

var (
	ErrBadValue  = errors.New("bad syntax")
	ErrBadSyntax = errors.New("bad value")
)

type Error struct {
	err     error
	context string
}

func (e *Error) Error() string {
	return e.context
	return e.err.Error() + ": " + e.context
}

func (e *Error) Unwrap() error {
	return e.err
}

type ParserError struct {
	err     error
	line    string
	context string
	pos     Position
	length  int `default:"1"`
}

func (e *ParserError) Error() string {
	str := "Parsing Error: " + e.context + "\n"
	deli := " |"
	pre := strings.Repeat(" ", 4) + strconv.Itoa(e.pos.row)
	str += pre + deli + e.line + "\n"
	str += strings.Repeat(" ", len(pre)) + deli
	start := e.pos.col
	stop := e.pos.col + e.length
	if start > 0 {
		str += strings.Repeat(" ", start)
	}
	str += strings.Repeat("^", e.length)
	if stop < len(e.line) {
		str += strings.Repeat(" ", len(e.line)-stop)
	}
	return str
}

func (e *ParserError) Unwrap() error {
	return e.err
}
