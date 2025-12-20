package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/config"
	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/rabbitmq"
)

var (
	configFile string
	version    bool
)

func init() {
	flag.StringVar(&configFile, "config", "./configs/config.yaml", "Path to configuration file")
	flag.BoolVar(&version, "version", false, "Show version")
}

func main() {
	flag.Parse()

	if version {
		printVersion()
		return
	}

	conf, err := config.NewConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logg := logger.New(conf.Logger.Level)

	rmq, err := rabbitmq.NewClient(conf.RabbitMQ.URL, conf.RabbitMQ.Queue)
	if err != nil {
		logg.Error(fmt.Sprintf("Failed to connect to RabbitMQ: %v", err))
		return
	}
	defer rmq.Close()

	msgs, err := rmq.Consume()
	if err != nil {
		logg.Error(fmt.Sprintf("Failed to start consuming: %v", err))
		return
	}

	logg.Info("sender is running...")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	for {
		select {
		case <-ctx.Done():
			logg.Info("sender is stopping...")
			return
		case msg, ok := <-msgs:
			if !ok {
				logg.Info("message channel closed")
				return
			}
			var n rabbitmq.Notification
			if err := json.Unmarshal(msg.Body, &n); err != nil {
				logg.Error(fmt.Sprintf("failed to unmarshal notification: %v", err))
				continue
			}
			logg.Infof("Notification: EventID=%s, Title=%s, UserID=%s, StartTime=%v", n.EventID, n.Title, n.UserID, n.StartTime)
		}
	}
}
