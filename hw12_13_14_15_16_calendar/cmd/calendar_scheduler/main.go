package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/config"
	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/rabbitmq"
	"github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/scheduler"
	sqlstorage "github.com/stas-ik/otus-go-test/hw12_13_14_15_16_calendar/internal/storage/sql"
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

	// Подключение к БД
	stor := sqlstorage.New(conf.Database.DSN)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := stor.Connect(ctx); err != nil {
		logg.Error(fmt.Sprintf("Failed to connect to database: %v", err))
		cancel()
		return
	}
	cancel()
	defer stor.Close(context.Background())

	// Подключение к RabbitMQ
	rmq, err := rabbitmq.NewClient(conf.RabbitMQ.URL, conf.RabbitMQ.Queue)
	if err != nil {
		logg.Error(fmt.Sprintf("Failed to connect to RabbitMQ: %v", err))
		return
	}
	defer rmq.Close()

	sched := scheduler.New(stor, rmq, logg)

	logg.Info("scheduler is running...")

	tickerScan := time.NewTicker(conf.Schedule.ScanInterval)
	defer tickerScan.Stop()

	tickerCleanup := time.NewTicker(conf.Schedule.CleanupInterval)
	defer tickerCleanup.Stop()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-tickerScan.C:
			sched.ProcessNotifications(context.Background())
		case <-tickerCleanup.C:
			sched.ProcessCleanup(context.Background())
		case <-stopCh:
			logg.Info("scheduler is stopping...")
			return
		}
	}
}
