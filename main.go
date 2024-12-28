package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/malletgaetan/dockermon/internal/config"
	"github.com/malletgaetan/dockermon/internal/logger"
)

var (
	configFilePath = flag.String("c", "", "configuration file path")
	logsJSON       = flag.Bool("f", false, "logs in JSON")
)

func main() {
	runtime.GOMAXPROCS(1)
	var wg sync.WaitGroup
	var loggerConfig logger.Config

	flag.Parse()
	loggerConfig.JSON = *logsJSON
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	msgs, errs := cli.Events(ctx, types.EventsOptions{
		Filters: conf.Filters(),
	})

	for {
		select {
		case err := <-errs:
			if err != nil {
				logger.Log.Error("docker events subscribe failed", "err", err)
				stop()
				goto out
			}
		case msg := <-msgs:
			cmd := conf.GetCmd(string(msg.Type), string(msg.Action))
			logger.Log.Info("handling event", "type", string(msg.Type), "action", string(msg.Action))
			wg.Add(1)
			go cmd.Execute(&wg, msg)
		}
	}
out:

	wg.Wait()

	logger.Log.Info("graceful shutdown complete")
}
