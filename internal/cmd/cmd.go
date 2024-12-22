package cmd

import (
	"context"
	"encoding/json"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/malletgaetan/dockermon/internal/logger"
)

type Cmd struct {
	Name    string
	Args    []string
	Timeout uint
}

// intended to be executed in a Go Routine
func (c *Cmd) Execute(wg *sync.WaitGroup, msg events.Message) {
	defer wg.Done()

	var cmd *exec.Cmd

	if c.Timeout != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(c.Timeout))
		defer cancel()
		cmd = exec.CommandContext(ctx, c.Name, c.Args...)
	} else {
		cmd = exec.Command(c.Name, c.Args...)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		logger.Log.Error("failed to retrieve command stdin", "err", err)
		return
	}

	// don't make signals propagate by creating different PGID for command
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		logger.Log.Error("failed to start command", "err", err)
		return
	}

	if err := json.NewEncoder(stdin).Encode(msg); err != nil {
		logger.Log.Error("failed to encode event in stdin", "err", err)
		return
	}

	if err := cmd.Wait(); err != nil {
		logger.Log.Error("failed to wait for command", "err", err)
		return
	}
}
