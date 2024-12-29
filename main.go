package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/malletgaetan/dockermon/internal/config"
	"github.com/malletgaetan/dockermon/internal/logger"
)

var (
	configFilePath = flag.String("c", "", "configuration file path")
	dumpConfig     = flag.Bool("d", false, "dump parsed configuration")
)

func exponentialBackoff(fn func() error, maxRetries int) error {
	var err error
	i := 0

	for {
		err = fn()
		if err == nil {
			return nil
		}
		if i == maxRetries {
			break
		}
		i++
		logger.Log.Error("failed to retrieve docker events", "err", err, "try", i)

		backoffDuration := time.Duration(math.Pow(2, float64(i))) * 200 * time.Millisecond
		time.Sleep(backoffDuration)
	}

	return err
}

func handleEvents(client *client.Client, conf *config.Config) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	msgs, errs := client.Events(ctx, events.ListOptions{
		Filters: conf.Filters(),
	})

	for {
		select {
		case err := <-errs:
			if err != nil {
				if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
					return nil
				}
				return err
			}
		case msg := <-msgs:
			cmd, err := conf.GetCmd(string(msg.Type), string(msg.Action))
			if err != nil {
				logger.Log.Error("failed to handle event", "type", string(msg.Type), "action", string(msg.Action), "err", err)
				continue
			}
			logger.Log.Info("handling event", "type", string(msg.Type), "action", string(msg.Action))
			wg.Add(1)
			go cmd.Execute(&wg, msg)
		}
	}
}

func main() {
	runtime.GOMAXPROCS(1)
	var loggerConfig logger.Config

	flag.Parse()
	loggerConfig.Level = slog.LevelInfo
	logger.Initialize(loggerConfig)

	if *configFilePath == "" {
		fmt.Println("Error: configuration file is required")
		os.Exit(1)
	}

	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		fmt.Println("Error: failed to create docker client: ", err)
		os.Exit(1)
	}

	hints, err := config.Setup(cli.ClientVersion())
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	file, err := os.Open(*configFilePath)
	if err != nil {
		fmt.Println("Error: failed to open config file: ", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(file)

	conf, err := config.ParseConfig(scanner, hints)
	file.Close()
	if err != nil {
		fmt.Println("Error: failed to parse configuration")
		fmt.Println(err)
		os.Exit(1)
	}

	if *dumpConfig {
		conf.Dump()
	}

	err = exponentialBackoff(func() error { return handleEvents(cli, conf) }, 6)
	if err != nil {
		logger.Log.Error("failed to listen to docker events", "err", err)
	}

	logger.Log.Info("Shutdown complete")
}
